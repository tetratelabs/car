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
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/tetratelabs/car/internal"
)

type car struct {
	registry internal.Registry
	out      io.Writer
	// patterns just like tar. Ex "car -tf image:tag foo/* bar.txt"
	patterns             []string
	verbose, veryVerbose bool
}

type patternMatcher struct {
	patterns map[string]bool
}

func newPatternMatcher(patterns []string) *patternMatcher {
	pm := &patternMatcher{patterns: map[string]bool{}}
	for _, pattern := range patterns {
		pm.patterns[pattern] = false
	}
	return pm
}

func (pm *patternMatcher) matchesPattern(name string) bool {
	matched := len(pm.patterns) == 0
	for pattern := range pm.patterns {
		if ok, _ := filepath.Match(pattern, name); ok {
			pm.patterns[pattern] = true
			matched = true
		}
	}
	return matched
}

func (pm *patternMatcher) unmatched() []string {
	unmatched := make([]string, 0, len(pm.patterns))
	for pattern, matched := range pm.patterns {
		if !matched {
			unmatched = append(unmatched, pattern)
		}
	}
	return unmatched
}

func (c *car) list(ctx context.Context, tag, platform string) error {
	img, err := c.registry.GetImage(ctx, tag, platform)
	if err != nil {
		return err
	}
	if c.veryVerbose {
		fmt.Fprintln(c.out, img.String()) //nolint
	}

	pm := newPatternMatcher(c.patterns)
	rf := c.listFunction(pm)

	for _, layer := range img.FilesystemLayers {
		if c.veryVerbose {
			fmt.Fprintln(c.out, layer.String()) //nolint
		}
		if err := c.registry.ReadFilesystemLayer(ctx, layer, rf); err != nil {
			return err
		}
	}

	unmatched := pm.unmatched()
	if len(unmatched) > 0 {
		return fmt.Errorf("%s not found in layer", strings.Join(unmatched, ", "))
	}
	return nil
}

func (c *car) listFunction(pm *patternMatcher) internal.ReadFile {
	return func(name string, size, mode int64, modTime time.Time, _ io.Reader) error {
		if !pm.matchesPattern(name) {
			return nil
		}
		if c.verbose || c.veryVerbose {
			fmt.Fprintf(c.out, "%s\t%d\t%s\t%s\n", fs.FileMode(mode), size, modTime.Format(time.Stamp), name)
		} else {
			fmt.Fprintln(c.out, name)
		}
		return nil
	}
}
