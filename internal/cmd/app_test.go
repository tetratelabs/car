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
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/registry/fake"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedStatus int
		expectedStdout string
		expectedStderr string
	}{
		{
			name:           "incorrect flag name",
			args:           []string{"car", "--d"},
			expectedStatus: 1,
			expectedStderr: `flag provided but not defined: -d
show usage with: car help
`,
		},
		{
			name:           "missing file value",
			args:           []string{"car", "-tf"},
			expectedStatus: 1,
			expectedStderr: `flag needs an argument: -f
show usage with: car help
`,
		},
		{
			name:           "incorrect file value",
			args:           []string{"car", "-tf", "icecream"},
			expectedStatus: 1,
			expectedStderr: `invalid [reference] flag: expected tagged reference
show usage with: car help
`,
		},
		{
			name:           "incorrect platform value",
			args:           []string{"car", "--platform", "icecream", "-tf", "tetratelabs/car:v1.0"},
			expectedStatus: 1,
			expectedStderr: `invalid [platform] flag: "icecream" should be 2 / delimited fields
show usage with: car help
`,
		},
		{
			name: "list",
			args: []string{"car", "-tf", "tetratelabs/car:v1.0"},
			expectedStdout: `bin/apple.txt
usr/local/bin/boat
usr/local/bin/car
Files/ProgramData/truck/bin/truck.exe
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
			name: "list matches layer-pattern",
			args: []string{"car", "--layer-pattern", "ADD", "-tf", "tetratelabs/car:v1.0", "usr/local/bin/*"},
			expectedStdout: `usr/local/bin/boat
usr/local/bin/car
`,
		},
		{
			name:           "list doesn't match layer-pattern",
			args:           []string{"car", "--layer-pattern", "/bin/sh", "-tf", "tetratelabs/car:v1.0", "usr/local/bin/car"},
			expectedStatus: 1,
			expectedStdout: ``,
			expectedStderr: `error: usr/local/bin/car not found in layer
`,
		},
	}

	for _, test := range tests {
		test := test // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(test.name, func(t *testing.T) {
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)

			status := Run(context.Background(), fake.NewRegistry, stdout, stderr, test.args)
			require.Equal(t, test.expectedStatus, status)
			require.Equal(t, test.expectedStdout, stdout.String())
			require.Equal(t, test.expectedStderr, stderr.String())
		})
	}
}
