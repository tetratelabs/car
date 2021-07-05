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

package car

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tetratelabs/car/internal"
	"github.com/tetratelabs/car/internal/patternmatcher"
)

// Car is like tar, except for containers.
type Car interface {
	// List prints any non-filtered files from the image layers of the given tag and platform.
	List(ctx context.Context, tag, platform string) error
	// Extract writes any non-filtered files from the image layers of the given tag and platform into the directory.
	// * directory must be absolute, though may be absent
	//
	// stripComponents strips the base directory of each internal.ReadFile call by the associated count.
	//   Ex directory=v1.0, stripComponents=1, name=/usr/bin/tar -> v1.0/bin/tar
	//   Ex directory=v1.0, stripComponents=2, name=/usr/bin/tar -> v1.0/tar
	//   Ex directory=v1.0, stripComponents=4, name=/usr/bin/tar -> ignored because too many path components
	Extract(ctx context.Context, tag, platform, directory string, stripComponents int) error
}

type car struct {
	registry         internal.Registry
	out              io.Writer
	createdByPattern *regexp.Regexp
	// filePatterns just like tar. Ex "car -tf image:tag foo/* bar.txt"
	filePatterns                   []string
	fastRead, verbose, veryVerbose bool
}

// New creates a new instance of Car
func New(registry internal.Registry, out io.Writer, createdByPattern *regexp.Regexp, patterns []string, fastRead, verbose, veryVerbose bool) Car {
	return &car{
		registry:         registry,
		out:              out,
		createdByPattern: createdByPattern,
		filePatterns:     patterns,
		fastRead:         fastRead,
		verbose:          verbose || veryVerbose,
		veryVerbose:      veryVerbose,
	}
}

func (c *car) do(ctx context.Context, readFile internal.ReadFile, tag, platform string) error {
	filteredLayers, err := c.getFilesystemLayers(ctx, tag, platform)
	if err != nil {
		return err
	}
	pm := patternmatcher.New(c.filePatterns, c.fastRead)
	rf := func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
		if !pm.MatchesPattern(name) {
			return nil
		}
		return readFile(name, size, mode, modTime, reader)
	}
	for _, layer := range filteredLayers {
		if c.veryVerbose {
			fmt.Fprintln(c.out, layer.String()) //nolint
		}
		if err := c.registry.ReadFilesystemLayer(ctx, layer, rf); err != nil {
			return err
		}
		if !pm.StillMatching() {
			break
		}
	}
	unmatched := pm.Unmatched()
	if len(unmatched) > 0 {
		return fmt.Errorf("%s not found in layer", strings.Join(unmatched, ", "))
	}
	return nil
}

func (c *car) Extract(ctx context.Context, tag, platform, directory string, stripComponents int) error {
	// maintain a lazy map of directories already created
	dirsCreated := map[string]struct{}{}
	return c.do(ctx, func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
		destinationPath, ok := newDestinationPath(name, directory, stripComponents)
		if !ok {
			return nil // skip
		}

		baseDir := path.Dir(destinationPath)
		if ok := dirsCreated[baseDir]; !ok {
			if err := os.MkdirAll(baseDir, 0755); err != nil { //nolint:gosec
				return err
			}
			dirsCreated[baseDir] = true
		}
		fw, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_RDWR, mode) //nolint:gosec
		if err != nil {
			return err
		}
		if c.veryVerbose {
			c.listVerbose(name, size, mode, modTime)
		} else if c.verbose {
			fmt.Fprintln(c.out, name)
		}
		_, err = io.CopyN(fw, reader, size)
		return err
	}, tag, platform)
}

func newDestinationPath(name, directory string, stripComponents int) (string, bool) {
	i := 0
	for ; stripComponents > 0 && i < len(name); i++ {
		if os.IsPathSeparator(name[i]) {
			stripComponents--
		}
	}
	// if the dirname length is longer than strip components, skip
	if stripComponents > 0 {
		return "", false
	}
	return filepath.Join(directory, name[i:]), true
}

func (c *car) List(ctx context.Context, tag, platform string) error {
	return c.do(ctx, func(name string, size int64, mode os.FileMode, modTime time.Time, _ io.Reader) error {
		if c.verbose {
			c.listVerbose(name, size, mode, modTime)
		} else {
			fmt.Fprintln(c.out, name)
		}
		return nil
	}, tag, platform)
}

func (c *car) listVerbose(name string, size int64, mode os.FileMode, modTime time.Time) {
	fmt.Fprintf(c.out, "%s\t%d\t%s\t%s\n", mode, size, modTime.Format(time.Stamp), name) //nolint
}

func (c *car) getFilesystemLayers(ctx context.Context, tag, platform string) ([]*internal.FilesystemLayer, error) {
	img, err := c.registry.GetImage(ctx, tag, platform)
	if err != nil {
		return nil, err
	}
	if c.veryVerbose {
		fmt.Fprintln(c.out, img.String()) //nolint
	}
	filteredLayers := make([]*internal.FilesystemLayer, 0, len(img.FilesystemLayers))
	for _, layer := range img.FilesystemLayers {
		if c.createdByPattern == nil || c.createdByPattern.MatchString(layer.CreatedBy) {
			filteredLayers = append(filteredLayers, layer)
		}
	}
	return filteredLayers, nil
}
