// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package car

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tetratelabs/car/api"
	"github.com/tetratelabs/car/internal"
	"github.com/tetratelabs/car/internal/patternmatcher"
)

// Car is like tar, except for containers.
//
// # Notes
//
//   - This is an interface for decoupling, not third-party implementations.
//     All implementations are in car.
type Car interface {
	internal.CarOnly

	// List prints any non-filtered files from the image layers of the given tag and platform.
	List(ctx context.Context, ref api.Reference, platform string) error

	// Extract writes any non-filtered files from the image layers of the given tag and platform into the directory.
	// * directory must be absolute, though may be absent
	//
	// stripComponents strips the base directory of each internal.ReadFile call by the associated count.
	//   Ex directory=v1.0, stripComponents=1, name=/usr/bin/tar -> v1.0/bin/tar
	//   Ex directory=v1.0, stripComponents=2, name=/usr/bin/tar -> v1.0/tar
	//   Ex directory=v1.0, stripComponents=4, name=/usr/bin/tar -> ignored because too many path components
	Extract(ctx context.Context, ref api.Reference, platform, directory string, stripComponents int) error
}

type car struct {
	internal.CarOnly

	registry         api.Registry
	out              io.Writer
	createdByPattern *regexp.Regexp
	// filePatterns just like tar. Ex "car -tf image:tag foo/* bar.txt"
	filePatterns                   []string
	fastRead, verbose, veryVerbose bool
}

// New creates a new instance of Car
func New(registry api.Registry, out io.Writer, createdByPattern *regexp.Regexp, patterns []string, fastRead, verbose, veryVerbose bool) Car {
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

func (c *car) do(ctx context.Context, readFile api.ReadFile, ref api.Reference, platform string) error {
	filteredLayers, err := c.getFilesystemLayers(ctx, ref, platform)
	if err != nil {
		return err
	}
	pm := patternmatcher.New(c.filePatterns, c.fastRead)
	rf := func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
		name = stripLeadingSlash(name)
		if !pm.MatchesPattern(name) {
			return nil
		}
		return readFile(name, size, mode, modTime, reader)
	}
	for _, layer := range filteredLayers {
		if c.veryVerbose {
			fmt.Fprintln(c.out, layer)
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

func (c *car) Extract(ctx context.Context, ref api.Reference, platform, directory string, stripComponents int) error {
	// maintain a lazy map of directories already created
	dirsCreated := map[string]struct{}{}
	return c.do(ctx, func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
		destinationPath, ok := newDestinationPath(name, directory, stripComponents)
		if !ok {
			return nil // skip
		}

		baseDir := filepath.Dir(destinationPath)
		if _, ok := dirsCreated[baseDir]; !ok {
			if err := os.MkdirAll(baseDir, 0o755); err != nil { //nolint:gosec // extraction needs predictable directory permissions
				return err
			}
			dirsCreated[baseDir] = struct{}{}
		}
		fw, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_RDWR, mode) //nolint:gosec // extraction preserves archive file modes
		if err != nil {
			return err
		}

		if c.veryVerbose { // extract veryVerbose = list verbose. In other words, tar -xvv output is the same as tar -tv
			c.listVerbose(name, size, mode, modTime)
		} else if c.verbose {
			fmt.Fprintln(c.out, name)
		}
		_, err = io.CopyN(fw, reader, size)
		return err
	}, ref, platform)
}

// newDestinationPath allows manipulation of the output path based on flags like `--strip-components`
// This returns the output path and a boolean which indicates if the file should be skipped or not.
func newDestinationPath(name, directory string, stripComponents int) (string, bool) {
	i := 0
	for ; stripComponents > 0 && i < len(name); i++ {
		if os.IsPathSeparator(name[i]) {
			stripComponents--
		}
	}
	// If the dirname length is longer than strip components, skip. We don't warn because tar doesn't, probably because
	// there could be many files skipped.
	if stripComponents > 0 {
		return "", false
	}
	return filepath.Join(directory, name[i:]), true
}

func (c *car) List(ctx context.Context, ref api.Reference, platform string) error {
	return c.do(ctx, func(name string, size int64, mode os.FileMode, modTime time.Time, _ io.Reader) error {
		if c.verbose {
			c.listVerbose(name, size, mode, modTime)
		} else {
			fmt.Fprintln(c.out, name)
		}
		return nil
	}, ref, platform)
}

func (c *car) listVerbose(name string, size int64, mode os.FileMode, modTime time.Time) {
	fmt.Fprintf(c.out, "%s\t%d\t%s\t%s\n", mode, size, modTime.Format(time.Stamp), name)
}

func (c *car) getFilesystemLayers(ctx context.Context, ref api.Reference, platform string) ([]api.FilesystemLayer, error) {
	img, err := c.registry.GetImage(ctx, ref, platform)
	if err != nil {
		return nil, err
	}
	if c.veryVerbose {
		fmt.Fprintln(c.out, img)
	}

	count := img.FilesystemLayerCount()
	filteredLayers := make([]api.FilesystemLayer, 0, img.FilesystemLayerCount())
	for i := range count {
		layer := img.FilesystemLayer(i)
		if c.createdByPattern == nil || c.createdByPattern.MatchString(layer.CreatedBy()) {
			filteredLayers = append(filteredLayers, layer)
		}
	}
	return filteredLayers, nil
}

// stripLeadingSlash removes any leading slash from the input file name, to
// normalize pattern matching. For example, paketo images have a combination of
// relative and absolute paths in their squashed image.
func stripLeadingSlash(name string) string {
	if name != "" && name[0] == '/' {
		return name[1:]
	}
	return name
}
