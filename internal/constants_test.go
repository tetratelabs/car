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

package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidArch(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{ArchAmd64, true},
		{ArchArm64, true},
		{"s390x", false},
		{"ice cream", false},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, IsValidArch(tc.name))
		})
	}
}

func TestIsValidOS(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{OSDarwin, true},
		{OSLinux, true},
		{OSWindows, true},
		{"solaris", false},
		{"ice cream", false},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, IsValidOS(tc.name))
		})
	}
}
