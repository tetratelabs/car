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

package main

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_unBundleFlags(t *testing.T) {
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
		{
			name:     "-t --platform linux/amd64 -qvvf",
			input:    []string{"-t", "--platform", "linux/amd64", "-qvvf"},
			expected: []string{"-t", "--platform", "linux/amd64", "-vv", "-q", "-f"},
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, unBundleFlags(tc.input))
		})
	}
}

func Test_platformValue(t *testing.T) {
	tests := []struct{ name, expectedErr string }{
		{name: "darwin/amd64"},
		{name: "darwin/arm64"},
		{name: "linux/amd64"},
		{name: "linux/arm64"},
		{name: "windows/amd64"},
		{name: "windows/arm64"},
		{
			name:        "darwin",
			expectedErr: `should be 2 / delimited fields`,
		},
		{
			name:        "darwin/amd64/11.3",
			expectedErr: `should be 2 / delimited fields`,
		},
		{
			name:        "solaris/amd64",
			expectedErr: `invalid OS`,
		},
		{
			name:        "windows/s390x",
			expectedErr: `invalid architecture`,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			var p platformValue
			err := p.Set(tc.name)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.name, string(p))
			}
		})
	}
}

func Test_referenceValue(t *testing.T) {
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
			var r referenceValue
			err := r.Set(tc.reference)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				domain, path, tag := r.Get()
				require.Equal(t, tc.expectedDomain, domain)
				require.Equal(t, tc.expectedPath, path)
				require.Equal(t, tc.expectedTag, tag)
			}
		})
	}
}

func Test_createdByPatternValue(t *testing.T) {
	tests := []struct {
		name            string
		expectedPattern *regexp.Regexp
		expectedErr     string
	}{
		{name: ``},
		{name: `ADD`, expectedPattern: regexp.MustCompile(`ADD`)},
		{name: `ADD.*envoy`, expectedPattern: regexp.MustCompile(`ADD.*envoy`)},
		{name: `(`, expectedErr: "error parsing regexp: missing closing ): `(`"},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			var c createdByPatternValue
			err := c.Set(tc.name)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPattern, c.p)
			}
		})
	}
}

func Test_directoryValue(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	tests := []struct{ name, expected string }{
		{name: "", expected: wd},
		{name: ".", expected: wd},
		{name: "foo", expected: filepath.Join(wd, "foo")},
		{name: "/foo", expected: "/foo"},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			var d directoryValue
			err := d.Set(tc.name)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(d))
		})
	}
}
