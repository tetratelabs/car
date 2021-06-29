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

	httpclient "github.com/tetratelabs/car/internal/httpclient"
)

func TestRoundTripper(t *testing.T) {
	type tagList struct {
		Name string
		Tags []string
	}
	expectedTagList := tagList{"envoy", []string{"v1.18.1", "v1.18.2"}}

	url, err := urlpkg.Parse("https://docker.io/v2/envoyproxy/envoy/tags/list?n=100")
	require.NoError(t, err)

	tests := []struct {
		name        string
		expectedErr string
		docker      http.RoundTripper
		real        http.RoundTripper
	}{
		{
			name:   "new",
			docker: NewRoundTripper("envoyproxy/envoy"),
			real: &mock{t, 0, []string{`GET /token?service=registry.docker.io&scope=repository:envoyproxy/envoy:pull HTTP/1.1
Host: auth.docker.io
User-Agent: car/dev
Accept: application/json

`, `GET /v2/envoyproxy/envoy/tags/list?n=100 HTTP/1.1
Host: docker.io
User-Agent: Go-http-client/1.1
Authorization: Bearer a

`}, []interface{}{tokenResponse{"a"}, expectedTagList}},
		},
		{
			name:   "valid",
			docker: &bearerAuth{"envoyproxy/envoy", "a"},
			real: &mock{t, 0, []string{`GET /v2/envoyproxy/envoy/tags/list?n=100 HTTP/1.1
Host: docker.io
User-Agent: Go-http-client/1.1
Authorization: Bearer a

`}, []interface{}{expectedTagList}},
		},
		{
			name:        "error",
			expectedErr: `received 401 status code from "https://auth.docker.io/token?service=registry.docker.io&scope=repository:envoyproxy/envoy:pull"`,
			docker:      &bearerAuth{"envoyproxy/envoy", ""},
			real:        &errMock{},
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := httpclient.ContextWithTransport(context.Background(), tc.real)
			req := &http.Request{Method: "GET", URL: url, Header: http.Header{}}
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
	return &http.Response{Status: "200 OK", StatusCode: 200,
		Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(bytes.NewReader(b))}, nil
}

type errMock struct{}

func (m *errMock) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{Status: "401 Unauthorized", StatusCode: 401}, nil
}
