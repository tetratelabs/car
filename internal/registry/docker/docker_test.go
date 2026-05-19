// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"encoding/json"
	"net/http"
	urlpkg "net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/httpclient"
	"github.com/tetratelabs/car/internal/test/httptest"
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
		handler     http.Handler
	}{
		{
			name:   "new",
			url:    url,
			docker: NewRoundTripper(),
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Host {
				case "auth.docker.io":
					assert.Equal(t, "/token", r.URL.Path)
					assert.Equal(t, "service=registry.docker.io&scope=repository:envoyproxy/envoy:pull", r.URL.RawQuery)
					assert.Equal(t, "application/json", r.Header.Get("Accept"))
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tokenResponse{"a"})
				case "index.docker.io":
					assert.Equal(t, "/v2/envoyproxy/envoy/manifests/list", r.URL.Path)
					assert.Equal(t, "n=100", r.URL.RawQuery)
					assert.Equal(t, "Bearer a", r.Header.Get("Authorization"))
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(expectedTagList)
				default:
					http.Error(w, "unexpected host: "+r.URL.Host, http.StatusInternalServerError)
				}
			}),
		},
		{
			name:   "valid",
			url:    url,
			docker: &bearerAuth{token: "a"},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "index.docker.io", r.URL.Host)
				assert.Equal(t, "/v2/envoyproxy/envoy/manifests/list", r.URL.Path)
				assert.Equal(t, "Bearer a", r.Header.Get("Authorization"))
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(expectedTagList)
			}),
		},
		{
			name:        "error",
			url:         url,
			expectedErr: `received 401 status code from "https://auth.docker.io/token?service=registry.docker.io&scope=repository:envoyproxy/envoy:pull"`,
			docker:      &bearerAuth{token: ""},
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			}),
		},
		{
			name:   "r2.cloudflarestorage.com",
			url:    r2URL,
			docker: &bearerAuth{token: "a"},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "docker-images-prod.6aa.r2.cloudflarestorage.com", r.URL.Host)
				assert.Empty(t, r.Header.Get("Authorization"), "Authorization header must not be sent to r2.cloudflarestorage.com")
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(expectedImageConfig)
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := httptest.HTTPClient(tc.handler)
			ctx := httpclient.ContextWithTransport(t.Context(), client.Transport)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, tc.url.String(), http.NoBody)
			require.NoError(t, err)
			res, err := tc.docker.RoundTrip(req)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				res.Body.Close()
			}
		})
	}
}
