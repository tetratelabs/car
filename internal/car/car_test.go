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
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/reference"
	"github.com/tetratelabs/car/internal/registry/fake"
)

func TestList(t *testing.T) {
	ref := reference.MustParse("ghcr.io/tetratelabs/car:v1.0")
	platform := "linux/amd64"

	tests := []struct {
		name                           string
		patterns                       []string
		createdByPattern               *regexp.Regexp
		fastRead, verbose, veryVerbose bool
		expectedOut, expectedErr       string
	}{
		{
			name: "normal",
			expectedOut: `bin/apple.txt
usr/local/bin/boat
usr/local/bin/car
Files/ProgramData/truck/bin/truck.exe
usr/local/sbin/car
`,
		},
		{
			name:     "all patterns match",
			patterns: []string{"bin/apple.txt", "usr/local/bin/*", "Files/ProgramData/truck/bin/*"},
			expectedOut: `bin/apple.txt
usr/local/bin/boat
usr/local/bin/car
Files/ProgramData/truck/bin/truck.exe
`,
		},
		{
			name:     "one pattern matches",
			patterns: []string{"usr/local/bin/*", "/etc"},
			expectedOut: `usr/local/bin/boat
usr/local/bin/car
`,
			expectedErr: "/etc not found in layer",
		},
		{
			name:     "not fast match",
			patterns: []string{"usr/local/bin/*"},
			expectedOut: `usr/local/bin/boat
usr/local/bin/car
`,
		},
		{
			name:     "fast match",
			fastRead: true,
			patterns: []string{"usr/local/bin/*"},
			expectedOut: `usr/local/bin/boat
`,
		},
		{
			name:        "fast match, very verbose",
			fastRead:    true,
			veryVerbose: true,
			patterns:    []string{"usr/local/bin/car"},
			expectedOut: `linux/amd64
4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f
15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2
-rwxr-xr-x	30	May 12 03:53:29	usr/local/bin/car
`,
		},
		{
			name:             "layer pattern",
			createdByPattern: regexp.MustCompile(`ADD build`),
			expectedOut: `usr/local/bin/car
usr/local/sbin/car
`,
		},
		{
			name:    "verbose",
			verbose: true,
			expectedOut: `-rw-r-----	10	Jun  7 06:28:15	bin/apple.txt
-rwxr-xr-x	20	Apr 16 22:53:09	usr/local/bin/boat
-rwxr-xr-x	30	May 12 03:53:29	usr/local/bin/car
-rw-r--r--	40	May 12 03:53:15	Files/ProgramData/truck/bin/truck.exe
-rwxr-xr-x	50	May 12 03:53:29	usr/local/sbin/car
`,
		},
		{
			name:        "veryVerbose",
			veryVerbose: true,
			expectedOut: `linux/amd64
4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f
-rw-r-----	10	Jun  7 06:28:15	bin/apple.txt
-rwxr-xr-x	20	Apr 16 22:53:09	usr/local/bin/boat
15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2
-rwxr-xr-x	30	May 12 03:53:29	usr/local/bin/car
1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81
-rw-r--r--	40	May 12 03:53:15	Files/ProgramData/truck/bin/truck.exe
6d2d8da2960b0044c22730be087e6d7b197ab215d78f9090a3dff8cb7c40c241
-rwxr-xr-x	50	May 12 03:53:29	usr/local/sbin/car
`,
		},
	}

	for _, test := range tests {
		tc := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			var stdout bytes.Buffer

			c := New(
				fake.Registry,
				&stdout,
				tc.createdByPattern,
				tc.patterns,
				tc.fastRead,
				tc.verbose,
				tc.veryVerbose,
			)

			if err := c.List(ctx, ref, platform); tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				require.Equal(t, tc.expectedOut, stdout.String())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOut, stdout.String())
			}
		})
	}
}

func TestExtract(t *testing.T) {
	ref := reference.MustParse("ghcr.io/tetratelabs/car:v1.0")
	platform := "linux/amd64"
	allFilesToSizes := map[string]int64{
		"bin/apple.txt":                         10,
		"usr/local/bin/boat":                    20,
		"usr/local/bin/car":                     30,
		"Files/ProgramData/truck/bin/truck.exe": 40,
		"usr/local/sbin/car":                    50,
	}

	tests := []struct {
		name                           string
		patterns                       []string
		createdByPattern               *regexp.Regexp
		fastRead, verbose, veryVerbose bool
		stripComponents                int
		expectedFileToSizes            map[string]int64
		expectedOut, expectedErr       string
	}{
		{
			name:                "normal",
			expectedFileToSizes: allFilesToSizes,
		},
		{
			name: "all patterns match",
			patterns: []string{
				"bin/apple.txt",
				"usr/local/bin/*",
				"Files/ProgramData/truck/bin/*",
				"usr/local/sbin/car",
			},
			expectedFileToSizes: allFilesToSizes,
		},
		{
			name:     "one pattern matches",
			patterns: []string{"usr/local/bin/*", "/etc"},
			expectedFileToSizes: map[string]int64{
				"usr/local/bin/boat": 20,
				"usr/local/bin/car":  30,
			},
			expectedErr: "/etc not found in layer",
		},
		{
			name:     "not fast match",
			patterns: []string{"usr/local/bin/*"},
			expectedFileToSizes: map[string]int64{
				"usr/local/bin/boat": 20,
				"usr/local/bin/car":  30,
			},
		},
		{
			name:     "fast match",
			fastRead: true,
			patterns: []string{"usr/local/bin/*"},
			expectedFileToSizes: map[string]int64{
				"usr/local/bin/boat": 20,
			},
		},
		{
			name:        "fast match, very verbose",
			fastRead:    true,
			veryVerbose: true,
			patterns:    []string{"usr/local/bin/car"},
			expectedFileToSizes: map[string]int64{
				"usr/local/bin/car": 30,
			},
			expectedOut: `linux/amd64
4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f
15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2
-rwxr-xr-x	30	May 12 03:53:29	usr/local/bin/car
`,
		},
		{
			name:             "layer pattern",
			createdByPattern: regexp.MustCompile(`ADD build`),
			expectedFileToSizes: map[string]int64{
				"usr/local/bin/car":  30,
				"usr/local/sbin/car": 50,
			},
		},
		{
			name:            "strip components - same match overwrites",
			stripComponents: 3,
			verbose:         true,
			patterns:        []string{"usr/local/*/car"},
			expectedFileToSizes: map[string]int64{
				"car": 50, // overwrites
			},
			// Just like tar, the output is the names in the archive, not the destination names.
			// As output is streaming, you will see both input names even if stripping results in only one file.
			expectedOut: `usr/local/bin/car
usr/local/sbin/car
`,
		},
		{
			name:            "strip components - fastRead picks first",
			stripComponents: 3,
			fastRead:        true,
			patterns:        []string{"usr/local/*/car"},
			expectedFileToSizes: map[string]int64{
				"car": 30, // quit at first match, and strips
			},
		},
		{
			name:                "verbose",
			verbose:             true,
			expectedFileToSizes: allFilesToSizes,
			expectedOut: `bin/apple.txt
usr/local/bin/boat
usr/local/bin/car
Files/ProgramData/truck/bin/truck.exe
usr/local/sbin/car
`,
		},
		{
			name:                "veryVerbose",
			veryVerbose:         true,
			expectedFileToSizes: allFilesToSizes,
			expectedOut: `linux/amd64
4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f
-rw-r-----	10	Jun  7 06:28:15	bin/apple.txt
-rwxr-xr-x	20	Apr 16 22:53:09	usr/local/bin/boat
15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2
-rwxr-xr-x	30	May 12 03:53:29	usr/local/bin/car
1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81
-rw-r--r--	40	May 12 03:53:15	Files/ProgramData/truck/bin/truck.exe
6d2d8da2960b0044c22730be087e6d7b197ab215d78f9090a3dff8cb7c40c241
-rwxr-xr-x	50	May 12 03:53:29	usr/local/sbin/car
`,
		},
	}

	for _, test := range tests {
		tc := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			var stdout bytes.Buffer
			c := New(
				fake.Registry,
				&stdout,
				tc.createdByPattern,
				tc.patterns,
				tc.fastRead,
				tc.verbose,
				tc.veryVerbose,
			)

			directory := t.TempDir()
			if err := c.Extract(ctx, ref, platform, directory, tc.stripComponents); tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
				require.Equal(t, tc.expectedOut, stdout.String())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOut, stdout.String())
			}
			for file, size := range tc.expectedFileToSizes {
				stat, err := os.Stat(filepath.Join(directory, file))
				require.NoError(t, err)
				require.True(t, !stat.IsDir())
				require.Equal(t, size, stat.Size())
			}
		})
	}
}

func TestNewDestinationPath(t *testing.T) {
	tests := []struct {
		name                      string
		inputName, inputDirectory string
		stripComponents           int
		expected                  string
		expectedOk                bool
	}{
		{
			name:           "base path",
			inputName:      "file",
			inputDirectory: "dir",
			expected:       "dir/file",
			expectedOk:     true,
		},
		{
			name:            "base path: can't strip",
			inputName:       "file",
			inputDirectory:  "dir",
			stripComponents: 1,
		},
		{
			name:           "one path",
			inputName:      "dir/file",
			inputDirectory: "dir",
			expected:       "dir/dir/file",
			expectedOk:     true,
		},
		{
			name:            "one path: strip one",
			inputName:       "dir/file",
			inputDirectory:  "dir",
			stripComponents: 1,
			expected:        "dir/file",
			expectedOk:      true,
		},
		{
			name:            "one path: can't strip two",
			inputName:       "dir/file",
			inputDirectory:  "dir",
			stripComponents: 2,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			have, ok := newDestinationPath(tc.inputName, tc.inputDirectory, tc.stripComponents)
			if !tc.expectedOk {
				require.False(t, ok)
			} else {
				require.True(t, ok)
				require.Equal(t, tc.expected, have)
			}
		})
	}
}
