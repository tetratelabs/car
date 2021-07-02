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

package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnBundleFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "empty",
		},
		{
			name:     "-q not bundled",
			input:    []string{"-q"},
			expected: []string{"-q"},
		},
		{
			name:     "-v not bundled",
			input:    []string{"-v"},
			expected: []string{"-v"},
		},
		{
			name:     "-vv not bundled",
			input:    []string{"-vv"},
			expected: []string{"-vv"},
		},
		{
			name:     "long flag left alone",
			input:    []string{"--tvvf"},
			expected: []string{"--tvvf"},
		},
		{
			name:     "non-flag left alone",
			input:    []string{"tvvf"},
			expected: []string{"tvvf"},
		},
		{
			name:     "not special flag left alone",
			input:    []string{"-f"},
			expected: []string{"-f"},
		},
		{
			name:     "-tvf",
			input:    []string{"-tvf"},
			expected: []string{"-v", "-t", "-f"},
		},
		{
			name:     "-tvvf",
			input:    []string{"-tvvf"},
			expected: []string{"-vv", "-t", "-f"},
		},
		{
			name:     "-qtvvf",
			input:    []string{"-qtvvf"},
			expected: []string{"-vv", "-q", "-t", "-f"},
		},
		{
			name:     "-tqvvf",
			input:    []string{"-tqvvf"},
			expected: []string{"-vv", "-q", "-t", "-f"},
		},
		{
			name:     "-xvf",
			input:    []string{"-xvf"},
			expected: []string{"-v", "-x", "-f"},
		},
		{
			name:     "-xvvf",
			input:    []string{"-xvvf"},
			expected: []string{"-vv", "-x", "-f"},
		},
		{
			name:     "--platform linux/amd64 -tvf",
			input:    []string{"--platform", "linux/amd64", "-tvf"},
			expected: []string{"--platform", "linux/amd64", "-v", "-t", "-f"},
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, unBundleFlags(tc.input))
		})
	}
}

func TestValidatePlatform(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr string
	}{
		{name: "darwin/amd64"},
		{name: "darwin/arm64"},
		{name: "linux/amd64"},
		{name: "linux/arm64"},
		{name: "windows/amd64"},
		{name: "windows/arm64"},
		{
			name:        "darwin",
			expectedErr: `invalid [platform] flag: "darwin" should be 2 / delimited fields`,
		},
		{
			name:        "darwin/amd64/11.3",
			expectedErr: `invalid [platform] flag: "darwin/amd64/11.3" should be 2 / delimited fields`,
		},
		{
			name:        "solaris/amd64",
			expectedErr: `invalid [platform] flag: "solaris/amd64" has an invalid OS`,
		},
		{
			name:        "windows/s390x",
			expectedErr: `invalid [platform] flag: "windows/s390x" has an invalid architecture`,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			platform, err := validatePlatformFlag(tc.name)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.name, platform)
			}
		})
	}
}

func TestValidateReference(t *testing.T) {
	tests := []struct{ name, reference, expectedDomain, expectedPath, expectedTag, expectedErr string }{
		{
			name:           "docker familiar",
			reference:      "envoyproxy/envoy:v1.18.3",
			expectedDomain: "docker.io",
			expectedPath:   "envoyproxy/envoy",
			expectedTag:    "v1.18.3",
		},
		{
			name:           "docker fully qualified",
			reference:      "docker.io/envoyproxy/envoy:v1.18.3",
			expectedDomain: "docker.io",
			expectedPath:   "envoyproxy/envoy",
			expectedTag:    "v1.18.3",
		},
		{
			name:           "docker familiar official",
			reference:      "alpine:3.14.0",
			expectedDomain: "docker.io",
			expectedPath:   "library/alpine",
			expectedTag:    "3.14.0",
		},
		{
			name:           "docker unfamiliar official",
			reference:      "docker.io/library/alpine:3.14.0",
			expectedDomain: "docker.io",
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
			name:           "port 5000 (ex. docker compose)",
			reference:      "registry:5000/tetratelabs/car:latest",
			expectedDomain: "registry:5000",
			expectedPath:   "tetratelabs/car",
			expectedTag:    "latest",
		},
		{
			name:        "missing tag",
			reference:   "registry:5000/tetratelabs/car",
			expectedErr: "invalid [reference] flag: expected tagged reference",
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			domain, path, tag, err := validateReferenceFlag(tc.reference)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedDomain, domain)
				require.Equal(t, tc.expectedPath, path)
				require.Equal(t, tc.expectedTag, tag)
			}
		})
	}
}
