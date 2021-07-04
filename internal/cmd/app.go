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
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/car/internal"
	carutil "github.com/tetratelabs/car/internal/car"
)

// validationError is arg marker of arg validation error vs an execution one.
type validationError struct {
	string
}

// Error implements the error interface.
func (e *validationError) Error() string {
	return e.string
}

// Run handles all error logging and coding so that no other place needs to.
func Run(ctx context.Context, newRegistry internal.NewRegistry, stdout, stderr io.Writer, args []string) int {
	argsToUse := unBundleFlags(args)

	app := newApp(newRegistry)
	app.Writer = stdout
	app.ErrWriter = stderr
	if err := app.RunContext(ctx, argsToUse); err != nil {
		// work around https://github.com/urfave/cli/pull/1285
		if len(argsToUse) == 1 || len(argsToUse) == 2 && argsToUse[1] == "help" {
			return 0
		}
		if _, ok := err.(*validationError); ok {
			fmt.Fprintln(stderr, err) //nolint
			logUsageError(app.Name, stderr)
		} else {
			fmt.Fprintln(stderr, "error:", err) //nolint
		}
		return 1
	}
	return 0
}

func logUsageError(name string, stderr io.Writer) {
	fmt.Fprintln(stderr, "show usage with:", name, "help") //nolint
}

func newApp(newRegistry internal.NewRegistry) *cli.App {
	var domain, path, tag, platform string
	var layerPattern *regexp.Regexp
	a := &cli.App{
		Name:     "car",
		HelpName: "car",
		Usage:    "car is like tar, but for containers!",
		Flags:    flags(),
		HideHelp: true,
		OnUsageError: func(c *cli.Context, err error, isSub bool) error {
			return &validationError{err.Error()}
		},
		Before: func(c *cli.Context) (err error) {
			domain, path, tag, err = validateReferenceFlag(c.String(flagReference))
			if err != nil {
				return err
			}
			platform, err = validatePlatformFlag(c.String(flagPlatform))
			if err != nil {
				return err
			}
			layerPattern, err = validateLayerPatternFlag(c.String(flagLayerPattern))
			return err
		},
		Action: func(c *cli.Context) error {
			car := carutil.New(
				newRegistry(c.Context, domain, path),
				c.App.Writer,
				layerPattern,
				c.Args().Slice(),
				c.Bool(flagFastRead),
				c.Bool(flagVerbose),
				c.Bool(flagVeryVerbose),
			)
			if c.Bool(flagList) {
				return car.List(c.Context, tag, platform)
			}
			return nil
		},
	}
	return a
}
