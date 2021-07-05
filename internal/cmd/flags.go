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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/urfave/cli/v2"

	"github.com/tetratelabs/car/internal"
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

// flags is a function instead of a var to avoid unit tests tainting each-other (cli.Flag contains state).
func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  flagCreatedByPattern,
			Usage: "regular expression to match the 'created_by' field of image layers",
		},
		&cli.StringFlag{
			Name:        flagDirectory,
			Aliases:     []string{"C"},
			DefaultText: ".",
			Usage:       fmt.Sprintf("Change to [%s] before extracting files", flagDirectory),
		},
		&cli.BoolFlag{
			Name:    flagExtract,
			Aliases: []string{"x"},
			Usage:   "Extract the image filesystem layers.",
		},
		&cli.BoolFlag{
			Name:    flagList,
			Aliases: []string{"t"},
			Usage:   "List image filesystem layers to stdout.",
		},
		&cli.StringFlag{
			Name:  flagPlatform,
			Usage: "Required when multi-architecture. Ex. linux/arm64, darwin/amd64 or windows/amd64",
		},
		&cli.StringFlag{
			Name:     flagReference,
			Aliases:  []string{"f"},
			Required: true,
			Usage:    "OCI reference to list or extract files from. Ex. envoyproxy/envoy:v1.18.3 or ghcr.io/homebrew/core/envoy:1.18.3-1",
		},
		&cli.IntFlag{
			Name:        flagStripComponents,
			DefaultText: "NUMBER",
			Usage:       "Strip NUMBER leading components from file names on extraction.",
		},
		&cli.BoolFlag{
			Name:    flagVerbose,
			Aliases: []string{"v"},
			Usage: "Produce verbose output. In extract mode, this will list each file name as it is extracted." +
				"In list mode, this produces output similar to ls.",
		},
		&cli.BoolFlag{
			Name:    flagFastRead,
			Aliases: []string{"q"},
			Usage:   "Extract or list only the first archive entry that matches each pattern or filename operand.",
		},
		&cli.BoolFlag{
			Name:    flagVeryVerbose,
			Aliases: []string{"vv"},
			Usage:   "Produce very verbose output. This produces arg header for each image layer and file details similar to ls.",
		},
	}
}

func validateCreatedByPatternFlag(createdByPattern string) (*regexp.Regexp, error) {
	if createdByPattern == "" {
		return nil, nil
	}

	p, err := regexp.Compile(createdByPattern)
	if err != nil {
		return nil, &validationError{fmt.Sprintf("invalid [%s] flag: %s", flagCreatedByPattern, err)}
	}
	return p, nil
}

func validateDirectoryFlag(directory string) (string, error) {
	if directory == "" || directory == "." {
		return os.Getwd()
	}
	d, err := filepath.Abs(directory)
	if err != nil {
		return "", &validationError{fmt.Sprintf("invalid [%s] flag: %s", flagDirectory, err)}
	}
	return d, nil
}

func validatePlatformFlag(platform string) (string, error) {
	if platform == "" { // optional
		return platform, nil
	}
	s := strings.Split(platform, "/")
	if len(s) != 2 {
		return "", &validationError{fmt.Sprintf("invalid [%s] flag: %q should be 2 / delimited fields", flagPlatform, platform)}
	}
	if !internal.IsValidOS(s[0]) {
		return "", &validationError{fmt.Sprintf("invalid [%s] flag: %q has an invalid OS", flagPlatform, platform)}
	}
	if !internal.IsValidArch(s[1]) {
		return "", &validationError{fmt.Sprintf("invalid [%s] flag: %q has an invalid architecture", flagPlatform, platform)}
	}
	return platform, nil
}

func validateReferenceFlag(ref string) (domain, path, tag string, err error) {
	name, err := reference.ParseNormalizedNamed(ref)
	if err != nil {
		return "", "", "", &validationError{err.Error()}
	}
	if _, ok := name.(reference.NamedTagged); !ok {
		return "", "", "", &validationError{fmt.Sprintf("invalid [%s] flag: expected tagged reference", flagReference)}
	}
	domain = reference.Domain(name)
	path = reference.Path(name)
	tag = name.(reference.NamedTagged).Tag()
	return
}

func validateStripComponentsFlag(stripComponents int) (int, error) {
	if stripComponents < 0 {
		return 0, &validationError{fmt.Sprintf("invalid [%s] flag: must be a whole number", flagStripComponents)}
	}
	return stripComponents, nil
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
