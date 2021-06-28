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

package internal

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Registry an an abstraction over a potentially remote OCI registry.
type Registry interface {
	// GetImage returns a summary of an image tag for a given platform, including its layers (FilesystemLayer).
	// An error is returned if there is no image manifest or configuration for the given platform.
	GetImage(ctx context.Context, tag, platform string) (Image, error)

	// ReadFilesystemLayer iterates over the files in the the "tar.gz" represented by a FilesystemLayer
	// The readFile function is called for each regular file. Returning an error from readFile will exit this function.
	ReadFilesystemLayer(ctx context.Context, layer FilesystemLayer, readFile ReadFile) error
}

// ReadFile is a callback for each selected file in the FilesystemLayer. This is only called on regular files, which
// means it  doesn't support tracking the directory that contains them. As this is usually backed by a tar file, it is
// possible the same name will be encountered more than once. It is also possible files are filtered out.
//
// Arguments correspond with tar.Header fields and are unaltered when this is backed by a tar.
// The reader argument optionally reads from the current file until io.EOF. Use the size argument to be more precise.
type ReadFile func(name string, size int64, mode int64, modTime time.Time, reader io.Reader) error

// Image represents filesystem layers that make up an image on a specific Platform, parsed from the OCI manifest and
// configuration.
// See https://github.com/opencontainers/image-spec/blob/master/manifest.md
// See https://github.com/opencontainers/image-spec/blob/master/config.md
type Image struct {
	// URL is the manifest URL to this image in its registry
	URL string
	// Platform encodes 'runtime.GOOS/runtime.GOARCH'. Ex "darwin/amd64"
	Platform string

	// FilesystemLayers are the filesystem layers of this image
	FilesystemLayers []FilesystemLayer
}

func (i *Image) String() string {
	var size int64
	for j := range i.FilesystemLayers {
		size += i.FilesystemLayers[j].Size
	}
	return fmt.Sprintf("%s platform=%s totalLayerSize: %d", i.URL, i.Platform, size)
}

// FilesystemLayer is a reference to a non-empty, downloadable "tar.gz" file
// See https://github.com/opencontainers/image-spec/blob/master/layer.md
type FilesystemLayer struct {
	// URL is the manifest URL to this filesystem layer in its registry
	// Ex. "sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f"
	URL string
	// MediaType is the "Accept" header used to retrieve this "tar.gz"
	// Ex. "application/vnd.oci.image.layer.v1.tar+gzip"
	MediaType string
	// Size are the compressed size in bytes of this "tar.gz"
	Size int64
	// CreatedBy when present is the (usually Dockerfile) command that created the layer
	// Note: When a shell command, it is possible it this tarball doesn't contain anything.
	CreatedBy string
}

func (f *FilesystemLayer) String() string {
	return fmt.Sprintf("%s size=%d CreatedBy: %s", f.URL, f.Size, f.CreatedBy)
}
