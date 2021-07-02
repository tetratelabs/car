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

package car

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/registry/fake"
)

func TestList(t *testing.T) {
	tag := "v1.0"
	platform := "linux/amd64"

	tests := []struct {
		name                     string
		patterns                 []string
		verbose, veryVerbose     bool
		expectedOut, expectedErr string
	}{
		{
			name: "normal",
			expectedOut: `bin/bash
usr/local/bin/boat
usr/local/bin/car
Files/ProgramData/truck/bin/truck.exe
`,
		},
		{
			name:     "all patterns match",
			patterns: []string{"bin/bash", "usr/local/bin/*", "Files/ProgramData/truck/bin/*"},
			expectedOut: `bin/bash
usr/local/bin/boat
usr/local/bin/car
Files/ProgramData/truck/bin/truck.exe
`,
		},
		{
			name:     "one patternmatcher match",
			patterns: []string{"usr/local/bin/*", "/etc"},
			expectedOut: `usr/local/bin/boat
usr/local/bin/car
`,
			expectedErr: "/etc not found in layer",
		},
		{
			name:    "verbose",
			verbose: true,
			expectedOut: `-rwxr-xr-x	1113504	Jun  7 06:28:15	bin/bash
-rwxr-xr-x	1000000	Apr 16 22:53:09	usr/local/bin/boat
-rwxr-xr-x	4000000	May 12 03:53:29	usr/local/bin/car
-rwxr-xr-x	8000000	May 12 03:53:15	Files/ProgramData/truck/bin/truck.exe
`,
		},
		{
			name:        "veryVerbose",
			veryVerbose: true,
			expectedOut: `fake://ghcr.io/v2/tetratelabs/car/manifests/v1.0 platform=linux/amd64 totalLayerSize: 32697009
fake://ghcr.io/v2/tetratelabs/car/blobs/sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f size=26697009
CreatedBy: /bin/sh -c #(nop) ADD file:d7fa3c26651f9204a5629287a1a9a6e7dc6a0bc6eb499e82c433c0c8f67ff46b in / 
-rwxr-xr-x	1113504	Jun  7 06:28:15	bin/bash
-rwxr-xr-x	1000000	Apr 16 22:53:09	usr/local/bin/boat
fake://ghcr.io/v2/tetratelabs/car/blobs/sha256:15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2 size=2000000
CreatedBy: ADD build/* /usr/local/bin/ # buildkit
-rwxr-xr-x	4000000	May 12 03:53:29	usr/local/bin/car
fake://ghcr.io/v2/tetratelabs/car/blobs/sha256:1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81 size=4000000
CreatedBy: cmd /S /C powershell iex(iwr -useb https://moretrucks.io/install.ps1)
-rwxr-xr-x	8000000	May 12 03:53:15	Files/ProgramData/truck/bin/truck.exe
`,
		},
	}

	for _, test := range tests {
		tc := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			stdout := new(bytes.Buffer)
			c := New(
				fake.NewRegistry(ctx, "ghcr.io", "tetratelabs/car"),
				stdout,
				tc.patterns,
				tc.verbose,
				tc.veryVerbose,
			)

			err := c.List(ctx, tag, platform)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				require.Equal(t, tc.expectedOut, stdout.String())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOut, stdout.String())
			}
		})
	}
}
