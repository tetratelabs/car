// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain arg copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/httpclient"
	"github.com/tetratelabs/car/internal/registry/docker"
	"github.com/tetratelabs/car/internal/registry/github"
)

func TestNew(t *testing.T) {
	tests := []struct{ name, host, path, expectedBaseURL string }{
		{
			name:            "docker familiar",
			host:            "",
			path:            "envoyproxy/envoy",
			expectedBaseURL: "https://index.docker.io/v2/envoyproxy/envoy",
		},
		{
			name:            "docker fully qualified",
			host:            "docker.io",
			path:            "envoyproxy/envoy",
			expectedBaseURL: "https://index.docker.io/v2/envoyproxy/envoy",
		},
		{
			name:            "docker familiar official",
			host:            "",
			path:            "alpine",
			expectedBaseURL: "https://index.docker.io/v2/library/alpine",
		},
		{
			name:            "docker unfamiliar official",
			host:            "docker.io",
			path:            "library/alpine",
			expectedBaseURL: "https://index.docker.io/v2/library/alpine",
		},
		{
			name:            "ghcr.io",
			host:            "ghcr.io",
			path:            "tetratelabs/car",
			expectedBaseURL: "https://ghcr.io/v2/tetratelabs/car",
		},
		{
			name:            "ghcr.io multiple slashes",
			host:            "ghcr.io",
			path:            "homebrew/core/envoy",
			expectedBaseURL: "https://ghcr.io/v2/homebrew/core/envoy",
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			r := New(ctx, tc.host, tc.path).(*registry)
			require.Equal(t, tc.expectedBaseURL, r.baseURL)
			require.NotNil(t, r.httpClient)
		})
	}
}

func TestHttpClientTransport(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		host     string
		expected http.RoundTripper
	}{
		{
			name:     "default nothing in context",
			ctx:      context.Background(),
			expected: http.DefaultTransport,
		},
		{
			name:     "default something in context",
			ctx:      httpclient.ContextWithTransport(context.Background(), github.NewRoundTripper()),
			expected: github.NewRoundTripper(),
		},
		{
			name:     "Docker",
			ctx:      context.Background(),
			host:     "index.docker.io",
			expected: docker.NewRoundTripper(""),
		},
		{
			name:     "GitHub",
			ctx:      context.Background(),
			host:     "ghcr.io",
			expected: github.NewRoundTripper(),
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			transport := httpClientTransport(tc.ctx, tc.host, "")
			require.IsType(t, tc.expected, transport)
		})
	}
}
