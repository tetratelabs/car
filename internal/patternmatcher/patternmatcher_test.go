// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

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
		t.Run(tc.name, func(t *testing.T) {
			pm := New(tc.patterns, true)
			for _, p := range tc.inputs {
				pm.MatchesPattern(p)
			}
			require.Equal(t, tc.expected, pm.StillMatching())
		})
	}
}
