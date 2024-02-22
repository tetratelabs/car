// Copyright 2023 Tetrate
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

package reference

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Parse(t *testing.T) {
	tests := []struct{ name, reference, expectedDomain, expectedPath, expectedTag, expectedErr string }{
		{
			name:           "docker familiar",
			reference:      "envoyproxy/envoy:v1.18.3",
			expectedDomain: "index.docker.io",
			expectedPath:   "envoyproxy/envoy",
			expectedTag:    "v1.18.3",
		},
		{
			name:           "not docker familiar",
			reference:      "webassembly.azurecr.io/hello-wasm:v1",
			expectedDomain: "webassembly.azurecr.io",
			expectedPath:   "hello-wasm",
			expectedTag:    "v1",
		},
		{
			name:           "docker fully qualified",
			reference:      "docker.io/envoyproxy/envoy:v1.18.3",
			expectedDomain: "index.docker.io",
			expectedPath:   "envoyproxy/envoy",
			expectedTag:    "v1.18.3",
		},
		{
			name:           "docker familiar official",
			reference:      "alpine:3.14.0",
			expectedDomain: "index.docker.io",
			expectedPath:   "library/alpine",
			expectedTag:    "3.14.0",
		},
		{
			name:           "docker unfamiliar official",
			reference:      "docker.io/library/alpine:3.14.0",
			expectedDomain: "index.docker.io",
			expectedPath:   "library/alpine",
			expectedTag:    "3.14.0",
		},
		{
			name:           "ghcr.io",
			reference:      "ghcr.io/tetratelabs/car:latest",
			expectedDomain: "ghcr.io",
			expectedPath:   "tetratelabs/car",
			expectedTag:    "latest",
		},
		{
			name:           "ghcr.io multiple slashes",
			reference:      "ghcr.io/homebrew/core/envoy:1.18.3-1",
			expectedDomain: "ghcr.io",
			expectedPath:   "homebrew/core/envoy",
			expectedTag:    "1.18.3-1",
		},
		{
			name:           "port 5443",
			reference:      "localhost:5443/tetratelabs/car:latest",
			expectedDomain: "localhost:5443",
			expectedPath:   "tetratelabs/car",
			expectedTag:    "latest",
		},
		{
			name:           "port 5000 (localhost)",
			reference:      "localhost:5000/tetratelabs/car:latest",
			expectedDomain: "localhost:5000",
			expectedPath:   "tetratelabs/car",
			expectedTag:    "latest",
		},
		{
			name:           "port 5000 (127.0.0.1)",
			reference:      "127.0.0.1:5000/tetratelabs/car:latest",
			expectedDomain: "127.0.0.1:5000",
			expectedPath:   "tetratelabs/car",
			expectedTag:    "latest",
		},
		{
			name:           "port 5000 (e.g. docker compose)",
			reference:      "registry:5000/tetratelabs/car:latest",
			expectedDomain: "registry:5000",
			expectedPath:   "tetratelabs/car",
			expectedTag:    "latest",
		},
		{
			name:        "empty",
			reference:   "",
			expectedErr: "invalid reference format",
		},
		{
			name:        "docker familiar, but no tag",
			reference:   "foo/bar",
			expectedErr: "expected tagged reference",
		},
		{
			name:        "missing tag",
			reference:   "registry:5000/tetratelabs/car",
			expectedErr: "expected tagged reference",
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			r, err := Parse(tc.reference)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedDomain, r.domain)
				require.Equal(t, tc.expectedPath, r.path)
				require.Equal(t, tc.expectedTag, r.tag)
			}
		})
	}
}
