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

package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/registry/fake"
)

func TestCarHelp(t *testing.T) {
	help, err := os.ReadFile(filepath.Join("testdata", "car_help.txt"))
	require.NoError(t, err)

	for _, tc := range [][]string{{"car"}, {"car", "help"}} {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(strings.Join(tc, " "), func(t *testing.T) {
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			status := Run(context.Background(), fake.NewRegistry, stdout, stderr, tc)
			require.Equal(t, 0, status)
			require.Equal(t, string(help), stdout.String())
			require.Empty(t, stderr)
		})
	}
}
