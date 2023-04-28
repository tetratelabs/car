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

package main

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/api"
	"github.com/tetratelabs/car/internal/registry/fake"
)

func Test_doMain(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStatus int
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "help",
			args:           []string{"car", "-h"},
			expectedStderr: usage,
		},
		{
			name:           "incorrect flag name",
			args:           []string{"car", "--d"},
			expectedStatus: 1,
			expectedStderr: "flag provided but not defined: -d\n" + usage,
		},
		{
			name:           "missing reference value",
			args:           []string{"car", "-tf"},
			expectedStatus: 1,
			expectedStderr: "flag needs an argument: -f\n" + usage,
		},
		{
			name:           "invalid reference value",
			args:           []string{"car", "-tf", "icecream"},
			expectedStatus: 1,
			expectedStderr: "invalid value \"icecream\" for flag -f: expected tagged reference\n" + usage,
		},
		{
			name:           "missing platform value",
			args:           []string{"car", "--platform"},
			expectedStatus: 1,
			expectedStderr: "flag needs an argument: -platform\n" + usage,
		},
		{
			name:           "invalid platform value",
			args:           []string{"car", "--platform", "icecream", "-tf", "tetratelabs/car:v1.0"},
			expectedStatus: 1,
			expectedStderr: "invalid value \"icecream\" for flag -platform: should be 2 / delimited fields\n" + usage,
		},
		{
			name:           "missing created-by-pattern value",
			args:           []string{"car", "--created-by-pattern"},
			expectedStatus: 1,
			expectedStderr: "flag needs an argument: -created-by-pattern\n" + usage,
		},
		{
			name:           "invalid created-by-pattern value",
			args:           []string{"car", "--created-by-pattern", "(", "-tf", "tetratelabs/car:v1.0"},
			expectedStatus: 1,
			expectedStderr: "invalid value \"(\" for flag -created-by-pattern: error parsing regexp: missing closing ): `(`\n" + usage,
		},
		{
			name:           "missing strip-components value",
			args:           []string{"car", "--strip-components"},
			expectedStatus: 1,
			expectedStderr: "flag needs an argument: -strip-components\n" + usage,
		},
		{
			name:           "invalid strip-components value",
			args:           []string{"car", "--strip-components", "-1", "-tf", "tetratelabs/car:v1.0"},
			expectedStatus: 1,
			expectedStderr: "invalid value \"-1\" for flag -strip-components: parse error\n" + usage,
		},
		{
			name:           "missing directory value",
			args:           []string{"car", "--directory"},
			expectedStatus: 1,
			expectedStderr: "flag needs an argument: -directory\n" + usage,
		},
		{
			name:           "list and extract",
			args:           []string{"car", "-t", "-xf", "tetratelabs/car:v1.0"},
			expectedStatus: 1,
			expectedStderr: "you cannot combine flags [list] and [extract]\n" + usage,
		},
		{
			name: "list",
			args: []string{"car", "-tf", "tetratelabs/car:v1.0"},
			expectedStdout: `bin/apple.txt
usr/local/bin/boat
usr/local/bin/car
Files/ProgramData/truck/bin/truck.exe
usr/local/sbin/car
`,
		},
		{
			name: "list matches pattern",
			args: []string{"car", "-tf", "tetratelabs/car:v1.0", "usr/local/bin/*"},
			expectedStdout: `usr/local/bin/boat
usr/local/bin/car
`,
		},
		{
			name:           "list doesn't match pattern",
			args:           []string{"car", "-tf", "tetratelabs/car:v1.0", "usr/local/bin/*", "robots"},
			expectedStatus: 1,
			expectedStdout: `usr/local/bin/boat
usr/local/bin/car
`,
			expectedStderr: `error: robots not found in layer
`,
		},
		{
			name: "list matches created-by-pattern",
			args: []string{"car", "--created-by-pattern", "ADD", "-tf", "tetratelabs/car:v1.0", "usr/local/bin/*"},
			expectedStdout: `usr/local/bin/boat
usr/local/bin/car
`,
		},
		{
			name:           "list doesn't match created-by-pattern",
			args:           []string{"car", "--created-by-pattern", "/bin/sh", "-tf", "tetratelabs/car:v1.0", "usr/local/bin/car"},
			expectedStatus: 1,
			expectedStdout: ``,
			expectedStderr: `error: usr/local/bin/car not found in layer
`,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			exitCode, stdout, stderr := runMain(t, "", tt.args)

			require.Equal(t, tt.expectedStderr, stderr)
			require.Equal(t, tt.expectedStdout, stdout)
			require.Equal(t, tt.expectedStatus, exitCode)
		})
	}
}

func runMain(t *testing.T, workdir string, args []string) (int, string, string) {
	t.Helper()

	// Use a workdir override if supplied.
	if workdir != "" {
		oldcwd, err := os.Getwd()
		require.NoError(t, err)

		require.NoError(t, os.Chdir(workdir))
		defer func() {
			require.NoError(t, os.Chdir(oldcwd))
		}()
	}

	oldArgs := os.Args
	t.Cleanup(func() {
		os.Args = oldArgs
	})
	os.Args = args

	var exitCode int
	var stdout, stderr bytes.Buffer
	var exited bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				exited = true
			}
		}()
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

		doMain(context.Background(), func(ctx context.Context, host string) (api.Registry, error) {
			return fake.Registry, nil
		}, &stdout, &stderr, func(code int) {
			exitCode = code
			panic(code) // to exit the func and set the exit status.
		})
	}()

	require.True(t, exited)

	return exitCode, stdout.String(), stderr.String()
}
