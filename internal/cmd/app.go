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
	"io/fs"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/car/internal"
	"github.com/tetratelabs/car/internal/registry"
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
func Run(ctx context.Context, stdout, stderr io.Writer, args []string) int {
	argsToUse := unBundleFlags(args)

	app := newApp()
	app.Writer = stdout
	app.ErrWriter = stderr
	if err := app.RunContext(ctx, argsToUse); err != nil {
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

func newApp() *cli.App {
	var ref reference.NamedTagged
	a := &cli.App{
		Name:     "car",
		Usage:    "car is like tar, but for containers!",
		Flags:    flags(),
		HideHelp: true,
		OnUsageError: func(c *cli.Context, err error, isSub bool) error {
			return &validationError{err.Error()}
		},
		Before: func(c *cli.Context) (err error) {
			name, err := reference.ParseNormalizedNamed(c.String(flagReference))
			if err != nil {
				return &validationError{err.Error()}
			}
			if nt, ok := name.(reference.NamedTagged); ok {
				ref = nt
			} else {
				return &validationError{fmt.Sprintf("invalid [%s] flag: expected tagged reference", flagReference)}
			}
			return validatePlatformFlag(c.String(flagPlatform))
		},
		Action: func(c *cli.Context) error {
			r := registry.New(ref)
			img, err := r.GetImage(c.Context, ref.Tag(), c.String(flagPlatform))
			if err != nil {
				return err
			}
			if c.Bool(flagVeryVerbose) {
				fmt.Fprintln(c.App.Writer, img.String()) //nolint
			}
			for _, layer := range img.FilesystemLayers {
				if c.Bool(flagVeryVerbose) {
					fmt.Fprintln(c.App.Writer, layer.String()) //nolint
				}
				if c.Bool(flagList) {
					verbose := c.Bool(flagVerbose) || c.Bool(flagVeryVerbose)
					return listFilesystemLayer(c, r, layer, verbose)
				}
			}
			return nil
		},
	}
	return a
}

func listFilesystemLayer(c *cli.Context, r internal.Registry, layer *internal.FilesystemLayer, verbose bool) error {
	w := c.App.Writer
	return r.ReadFilesystemLayer(c.Context, layer, func(name string, size int64, mode int64, modTime time.Time, _ io.Reader) error {
		if verbose {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", fs.FileMode(mode), size, modTime.Format(time.Stamp), name)
		} else {
			fmt.Fprintln(w, name)
		}
		return nil
	})
}
