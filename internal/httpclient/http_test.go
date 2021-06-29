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

package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttpClient_Get(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		header           http.Header
		expectedRequests []string
	}{
		{
			name: "GitHub release - Authorization: none",
			url:  "https://api.github.com/repos/envoyproxy/envoy/releases?per_page=100",
			expectedRequests: []string{`GET /repos/envoyproxy/envoy/releases?per_page=100 HTTP/1.1
Host: api.github.com
User-Agent: car/dev

`},
		},
		{
			name: "Homebrew bottle",
			url:  "https://ghcr.io/v2/homebrew/core/envoy/manifests/1.18.3-1",
			header: http.Header{
				"Accept":        []string{"application/vnd.oci.image.index.v1+json"},
				"Authorization": []string{"Bearer QQ=="}},
			expectedRequests: []string{`GET /v2/homebrew/core/envoy/manifests/1.18.3-1 HTTP/1.1
Host: ghcr.io
User-Agent: car/dev
Accept: application/vnd.oci.image.index.v1+json
Authorization: Bearer QQ==

`}},
		{
			name: "Docker registry",
			url:  "https://docker.io/v2/envoyproxy/envoy/manifests/v1.18.3",
			header: http.Header{
				"Accept": []string{
					"application/vnd.docker.distribution.manifest.list.v2+json",
					"application/vnd.docker.distribution.manifest.v2+json",
				},
				"Authorization": []string{"Bearer eyJhbGciOiJSUzI1NiIsInR5cC"}},
			expectedRequests: []string{`GET /v2/envoyproxy/envoy/manifests/v1.18.3 HTTP/1.1
Host: docker.io
User-Agent: car/dev
Accept: application/vnd.docker.distribution.manifest.list.v2+json
Accept: application/vnd.docker.distribution.manifest.v2+json
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cC

`},
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			r := recorder{}
			client := New(&r)

			_, _, err := client.Get(context.Background(), tc.url, &tc.header)
			require.NoError(t, err)

			for i, e := range tc.expectedRequests {
				require.Equal(t, e, r.requests[i])
			}
		})
	}
}

// TestHttpClient_Get_ErrorsOnBadRequest tests errors prior to the actual request
func TestHttpClient_Get_ErrorsOnBadRequest(t *testing.T) {
	r := recorder{}
	_, _, err := New(&r).Get(context.Background(), "https://api.github.com/\n", &http.Header{})
	require.Error(t, err)
	require.Empty(t, r.requests)
}

func TestHttpClient_Get_Body(t *testing.T) {
	expectedBody, expectedMediaType := `{"foo", "bar"}`, "application/json"
	r := recorder{responseBody: expectedBody, responseHeaders: map[string][]string{"Content-Type": {expectedMediaType}}}
	body, mediaType, err := New(&r).Get(context.Background(), "https://api.github.com/", &http.Header{})
	require.NoError(t, err)
	defer body.Close()

	require.Equal(t, expectedMediaType, mediaType)
	b, err := io.ReadAll(body)
	require.NoError(t, err)
	require.Equal(t, expectedBody, string(b))
}

// TestHttpClient_Get_StripsLongContentTypes so that we can use case statements on the resulting mediaType
func TestHttpClient_Get_MediaTypeStripsLongContentTypes(t *testing.T) {
	r := recorder{responseHeaders: map[string][]string{"Content-Type": {"application/json; charset=utf-8"}}}
	_, mediaType, err := New(&r).Get(context.Background(), "https://api.github.com/", &http.Header{})
	require.NoError(t, err)
	require.Equal(t, "application/json", mediaType)
}

func TestTransportFromContext(t *testing.T) {
	require.Equal(t, http.DefaultTransport, TransportFromContext(context.Background()))

	r := &recorder{}
	ctx := ContextWithTransport(context.Background(), r)
	require.Same(t, r, TransportFromContext(ctx))
}

type recorder struct {
	requests        []string
	responseHeaders map[string][]string
	responseBody    string
}

func (r *recorder) RoundTrip(req *http.Request) (*http.Response, error) {
	raw := new(bytes.Buffer)
	req.Write(raw) //nolint
	r.requests = append(r.requests, strings.ReplaceAll(raw.String(), "\r\n", "\n"))
	body := io.NopCloser(strings.NewReader(r.responseBody))
	return &http.Response{Status: "200 OK", StatusCode: 200, Header: r.responseHeaders, Body: body}, nil
}
