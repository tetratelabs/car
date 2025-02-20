// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/httpclient"
)

func TestRoundTripper(t *testing.T) {
	type tagList struct {
		Name string
		Tags []string
	}
	expectedTagList := tagList{"envoy", []string{"v1.18.1", "v1.18.2"}}

	type imageConfig struct {
		Architecture string `json:"architecture"`
	}
	expectedImageConfig := imageConfig{Architecture: "amd64"}

	url, err := urlpkg.Parse("https://index.docker.io/v2/envoyproxy/envoy/manifests/list?n=100")
	require.NoError(t, err)

	r2URL, err := urlpkg.Parse("https://docker-images-prod.6aa.r2.cloudflarestorage.com/registry-v2/docker/registry/v2/blobs/sha256/28/28b3")
	require.NoError(t, err)

	tests := []struct {
		name        string
		url         *urlpkg.URL
		expectedErr string
		docker      http.RoundTripper
		real        http.RoundTripper
	}{
		{
			name:   "new",
			url:    url,
			docker: NewRoundTripper(),
			real: &mock{t, 0, []string{`GET /token?service=registry.docker.io&scope=repository:envoyproxy/envoy:pull HTTP/1.1
Host: auth.docker.io
Accept: application/json

`, `GET /v2/envoyproxy/envoy/manifests/list?n=100 HTTP/1.1
Host: index.docker.io
Authorization: Bearer a

`}, []interface{}{tokenResponse{"a"}, expectedTagList}},
		},
		{
			name:   "valid",
			url:    url,
			docker: &bearerAuth{"a"},
			real: &mock{t, 0, []string{`GET /v2/envoyproxy/envoy/manifests/list?n=100 HTTP/1.1
Host: index.docker.io
Authorization: Bearer a

`}, []interface{}{expectedTagList}},
		},
		{
			name:        "error",
			url:         url,
			expectedErr: `received 401 status code from "https://auth.docker.io/token?service=registry.docker.io&scope=repository:envoyproxy/envoy:pull"`,
			docker:      &bearerAuth{""},
			real:        &errMock{},
		},
		{
			name: "r2.cloudflarestorage.com",
			url:  r2URL,
			// While we set the Authorization header here, we don't want it to be sent to r2.cloudflarestorage.com.
			docker: &bearerAuth{"a"},
			real: &mock{t, 0, []string{`GET /registry-v2/docker/registry/v2/blobs/sha256/28/28b3 HTTP/1.1
Host: docker-images-prod.6aa.r2.cloudflarestorage.com

`}, []interface{}{expectedImageConfig}},
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := httpclient.ContextWithTransport(context.Background(), tc.real)
			req := &http.Request{Method: http.MethodGet, URL: tc.url, Header: http.Header{}}
			res, err := tc.docker.RoundTrip(req.WithContext(ctx))
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				res.Body.Close()
			}
		})
	}
}

type mock struct {
	t             *testing.T
	i             int
	requests      []string
	jsonResponses []interface{}
}

func (m *mock) RoundTrip(req *http.Request) (*http.Response, error) {
	raw := new(bytes.Buffer)
	req.Write(raw) //nolint
	require.Equal(m.t, m.requests[m.i], strings.ReplaceAll(raw.String(), "\r\n", "\n"))

	b, err := json.Marshal(m.jsonResponses[m.i])
	require.NoError(m.t, err)
	m.i++
	return &http.Response{
		Status: "200 OK", StatusCode: http.StatusOK,
		Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(bytes.NewReader(b)),
	}, nil
}

type errMock struct{}

func (m *errMock) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{Status: "401 Unauthorized", StatusCode: http.StatusUnauthorized}, nil
}
