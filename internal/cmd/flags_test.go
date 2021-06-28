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
			name:     "-xvf",
			input:    []string{"-xvf"},
			expected: []string{"-v", "-x", "-f"},
		},
		{
			name:     "-xvvf",
			input:    []string{"-xvvf"},
			expected: []string{"-vv", "-x", "-f"},
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
			if tc.expectedErr != "" {
				require.EqualError(t, validatePlatformFlag(tc.name), tc.expectedErr)
			} else {
				require.NoError(t, validatePlatformFlag(tc.name))
			}
		})
	}
}
