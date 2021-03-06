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

package patternmatcher

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		input    string
		expected bool
	}{
		{
			name:     "no patterns",
			input:    "usr/local/bin/car",
			expected: true,
		},
		{
			name:     "no pattern matches",
			input:    "usr/local/bin/car",
			patterns: []string{"usr/local/sbin", "etc"},
		},
		{
			name:     "only pattern matches (exact)",
			input:    "usr/local/bin/car",
			patterns: []string{"usr/local/bin/car"},
			expected: true,
		},
		{
			name:     "only pattern matches (glob)",
			input:    "usr/local/bin/car",
			patterns: []string{"usr/local/bin/*"},
			expected: true,
		},
		{
			name:     "one pattern matches",
			input:    "usr/local/bin/car",
			patterns: []string{"usr/local/bin/*", "etc"},
			expected: true,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			pm := New(tc.patterns, false)
			require.Equal(t, tc.expected, pm.MatchesPattern(tc.input))
		})
	}
}

func TestStillMatching(t *testing.T) {
	tests := []struct {
		name             string
		patterns, inputs []string
		expected         bool
	}{
		{
			name:     "no patterns",
			inputs:   []string{"usr/local/bin/car"},
			expected: true,
		},
		{
			name:     "no pattern matches",
			patterns: []string{"usr/local/bin", "etc"},
			inputs:   []string{"usr/local/bin/car"},
			expected: true,
		},
		{
			name:     "only pattern matches (exact)",
			patterns: []string{"usr/local/bin/car"},
			inputs:   []string{"usr/local/bin/car"},
		},
		{
			name:     "only pattern matches (glob)",
			patterns: []string{"usr/local/bin/*"},
			inputs:   []string{"usr/local/bin/car"},
		},
		{
			name:     "one pattern matches",
			patterns: []string{"usr/local/bin/*", "etc"},
			inputs:   []string{"usr/local/bin/car"},
			expected: true,
		},
		{
			name:     "all patterns match",
			patterns: []string{"usr/local/bin/*", "usr/local/bin/car"},
			inputs:   []string{"usr/local/bin/car"},
			expected: true,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			pm := New(tc.patterns, true)
			for _, p := range tc.inputs {
				pm.MatchesPattern(p)
			}
			require.Equal(t, tc.expected, pm.StillMatching())
		})
	}
}
