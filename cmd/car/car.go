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
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tetratelabs/car"
	"github.com/tetratelabs/car/api"
	internalcar "github.com/tetratelabs/car/internal/car"
)

const (
	flagCreatedByPattern = "created-by-pattern"
	flagDirectory        = "directory"
	flagExtract          = "extract"
	flagFastRead         = "fast-read"
	flagList             = "list"
	flagPlatform         = "platform"
	flagReference        = "reference"
	flagStripComponents  = "strip-components"
	flagVerbose          = "verbose"
	flagVeryVerbose      = "very-verbose"
)

var usage = `NAME:
   car - car is like tar, but for containers!

USAGE:
   car [global options] [arguments...]

GLOBAL OPTIONS:
   --created-by-pattern value   regular expression to match the 'created_by' field of image layers
   --directory value, -C value  Change to [directory] before extracting files (default: .)
   --extract, -x                Extract the image filesystem layers. (default: false)
   --fast-read, -q              Extract or list only the first archive entry that matches each pattern or filename operand. (default: false)
   --list, -t                   List image filesystem layers to stdout. (default: false)
   --platform value             Required when multi-architecture. e.g. linux/arm64, darwin/amd64 or windows/amd64
   --reference value, -f value  OCI reference to list or extract files from. e.g. envoyproxy/envoy:v1.18.3 or ghcr.io/homebrew/core/envoy:1.18.3-1
   --strip-components value     Strip NUMBER leading components from file names on extraction. (default: NUMBER)
   --verbose, -v                Produce verbose output. In extract mode, this will list each file name as it is extracted.In list mode, this produces output similar to ls. (default: false)
   --very-verbose, --vv         Produce very verbose output. This produces arg header for each image layer and file details similar to ls. (default: false)

`

func main() {
	doMain(context.Background(), car.NewRegistry, os.Stdout, os.Stderr, os.Exit)
}

// doMain is separated out for the purpose of unit testing.
func doMain(
	ctx context.Context,
	newRegistry func(ctx context.Context, host string) (api.Registry, error),
	stdout, stderr io.Writer,
	exit func(code int),
) {
	flag := flag.NewFlagSet("car", flag.ContinueOnError)
	flag.Usage = func() {
		_, _ = stderr.Write([]byte(usage))
	}
	flag.SetOutput(stderr)

	var help bool
	flag.BoolVar(&help, "h", false, "print usage")

	createdByPattern := createdByPatternValue{}
	flag.Var(&createdByPattern, flagCreatedByPattern,
		"regular expression to match the 'created_by' field of image layers")

	var directory directoryValue
	for _, n := range []string{flagDirectory, "C"} {
		flag.Var(&directory, n,
			fmt.Sprintf("Change to [%s] before extracting files", flagDirectory))
	}

	var extract bool
	for _, n := range []string{flagExtract, "x"} {
		flag.BoolVar(&extract, n, false, "Extract the image filesystem layers.")
	}

	var fastRead bool
	for _, n := range []string{flagFastRead, "q"} {
		flag.BoolVar(&fastRead, n, false, "Extract or list only the first archive entry that matches each pattern or filename operand.")
	}

	var list bool
	for _, n := range []string{flagList, "t"} {
		flag.BoolVar(&list, n, false, "List image filesystem layers to stdout. (default: false).")
	}

	var platform platformValue
	flag.Var(&platform, flagPlatform,
		"Required when multi-architecture. e.g. linux/arm64, darwin/amd64 or windows/amd64")

	imageRef := referenceValue{}
	for _, n := range []string{flagReference, "f"} {
		flag.Var(&imageRef, n,
			"OCI reference to list or extract files from. e.g. envoyproxy/envoy:v1.18.3 or ghcr.io/homebrew/core/envoy:1.18.3-1")
	}

	var stripComponents uint
	flag.UintVar(&stripComponents, flagStripComponents, 0,
		"Strip NUMBER leading components from file names on extraction.")

	var verbose bool
	for _, n := range []string{flagVerbose, "v"} {
		flag.BoolVar(&verbose, n, false, "Produce verbose output. In extract mode, this will list each file name as it is extracted."+
			"In list mode, this produces output similar to ls.")
	}

	var veryVerbose bool
	for _, n := range []string{flagVeryVerbose, "vv"} {
		flag.BoolVar(&veryVerbose, n, false, "Produce very verbose output. This produces arg header for each image layer and file details similar to ls.")
	}

	if err := flag.Parse(unBundleFlags(os.Args[1:])); err != nil {
		exit(1) // usage would have already been printed
	} else if help || len(os.Args) == 1 {
		flag.Usage()
		exit(0)
	} else {
		createdByPattern := createdByPattern.p
		ref := imageRef.r

		r, err := newRegistry(ctx, ref.Domain())
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			exit(1)
		}

		car := internalcar.New(
			r,
			stdout,
			createdByPattern,
			flag.Args(),
			fastRead,
			verbose,
			veryVerbose,
		)

		if list {
			if extract {
				fmt.Fprintf(stderr, "you cannot combine flags [%s] and [%s]\n%s", flagList, flagExtract, usage)
				exit(1)
			}
			err = car.List(ctx, ref, string(platform))
		} else if extract {
			err = car.Extract(ctx, ref, string(platform), string(directory), int(stripComponents))
		}
		if err != nil {
			fmt.Fprintln(stderr, "error:", err)
			exit(1)
		} else {
			exit(0)
		}
	}
}

// unBundleFlags allows tar-like syntax like `car -tvvf ghcr.io/homebrew/core/envoy:1.18.3-1`
func unBundleFlags(args []string) []string {
	var result []string
	for _, a := range args {
		if !strings.HasPrefix(a, "-") || strings.HasPrefix(a, "--") {
			result = append(result, a)
			continue
		}
		a = unBundleFlag(a, "vv", &result)
		a = unBundleFlag(a, "v", &result)
		a = unBundleFlag(a, "q", &result)
		switch a {
		case "":
			continue
		case "-tf":
			result = append(result, "-t", "-f")
		case "-xf":
			result = append(result, "-x", "-f")
		default:
			result = append(result, a)
		}
	}
	return result
}

func unBundleFlag(argIn, flag string, args *[]string) string {
	switch {
	case argIn == "-"+flag:
		*args = append(*args, argIn)
		return ""
	case strings.Contains(argIn, flag): // flag exists in the middle or the end
		*args = append(*args, "-"+flag)
		return strings.Replace(argIn, flag, "", 1)
	default:
		return argIn
	}
}

type referenceValue struct {
	r api.Reference
}

// Set implements flag.Value
func (r *referenceValue) Set(val string) (err error) {
	r.r, err = car.ParseReference(val)
	return
}

func (r *referenceValue) String() string {
	if r.r == nil {
		return ""
	}
	return r.r.String()
}

type platformValue string

// Set implements flag.Value
func (p *platformValue) Set(val string) error {
	if val == "" { // optional
		return nil
	}
	s := strings.Split(val, "/")
	if len(s) != 2 {
		return errors.New("should be 2 / delimited fields")
	}
	*p = platformValue(val)
	return nil
}

func (p *platformValue) String() string {
	return string(*p)
}

type createdByPatternValue struct {
	p *regexp.Regexp
}

// Set implements flag.Value
func (c *createdByPatternValue) Set(val string) error {
	if val == "" { // optional
		return nil
	}
	if p, err := regexp.Compile(val); err != nil {
		return err
	} else {
		*c = createdByPatternValue{p: p}
	}
	return nil
}

func (c *createdByPatternValue) String() string {
	if c.p == nil {
		return ""
	}
	return c.p.String()
}

type directoryValue string

// Set implements flag.Value
func (d *directoryValue) Set(val string) (err error) {
	if val == "" {
		val = "."
	}
	if val, err = filepath.Abs(val); err != nil {
		return
	}
	*d = directoryValue(val)
	return nil
}

func (d *directoryValue) String() string {
	return string(*d)
}
