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

package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tetratelabs/car/internal"
)

const (
	MediaTypeOCIImageConfig   = "application/vnd.oci.image.config.v1+json"
	MediaTypeOCIImageIndex    = "application/vnd.oci.image.index.v1+json"
	MediaTypeOCIImageLayer    = "application/vnd.oci.image.layer.v1.tar+gzip"
	MediaTypeOCIImageManifest = "application/vnd.oci.image.manifest.v1+json"

	MediaTypeDockerContainerImage    = "application/vnd.docker.container.image.v1+json"
	MediaTypeDockerImageLayer        = "application/vnd.docker.image.rootfs.diff.tar.gzip"
	MediaTypeDockerImageForeignLayer = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip"
	MediaTypeDockerManifest          = "application/vnd.docker.distribution.manifest.v2+json"
	MediaTypeDockerManifestList      = "application/vnd.docker.distribution.manifest.list.v2+json"

	// MediaTypeUnknownImageConfig is set by oras when a config isn't explicitly specified.
	// See https://github.com/oras-project/oras-go/blob/96a37c2b359ac1305f70dc31b28c789688d77d0f/pack.go#L35
	MediaTypeUnknownImageConfig = "application/vnd.unknown.config.v1+json"

	// MediaTypeWasmImageConfig is from Solo's "WASM Artifact Image Specification"
	// See https://github.com/solo-io/wasm/commit/7389be1a694af80784d5a593a98e20fde34876f3
	MediaTypeWasmImageConfig = "application/vnd.module.wasm.config.v1+json"

	// MediaTypeWasmImageLayer is from Solo's "WASM Artifact Image Specification"
	// See https://github.com/solo-io/wasm/commit/7389be1a694af80784d5a593a98e20fde34876f3
	MediaTypeWasmImageLayer = "application/vnd.module.wasm.content.layer.v1+wasm"
)

// Reference is a parsed OCI reference.
//
// # Notes
//
//   - This is an interface for decoupling, not third-party implementations.
//     All implementations are in car.
type Reference interface {
	internal.CarOnly

	Domain() string
	Path() string
	Tag() string

	fmt.Stringer
}

// Registry is an abstraction over a potentially remote OCI registry.
type Registry interface {
	internal.CarOnly

	// GetImage returns a summary of an image tag for a given platform,
	// including its layers (FilesystemLayer).
	//
	// # Parameters
	//
	//   - path: the image path which must include at least one slash, possibly
	//     more than two. The only paths allowed to exclude a slash are DockerHub
	//     official images like "alpine"
	//   - platform: possibly empty Image.Platform qualifier.
	//
	// # Errors
	//
	//   - there is no image manifest
	//   - The platform parameter is empty, but there is more than one platform
	//     choice in the image.
	//   - The platform parameter does not match a platform in the image.
	GetImage(ctx context.Context, ref Reference, platform string) (Image, error)

	// ReadFilesystemLayer iterates over the files in the "tar.gz" represented
	// by a FilesystemLayer
	//
	// # Parameters
	//
	//   - layer: a chosen layer from Image.FilesystemLayers
	//   - readFile: a callback for each regular file.
	//
	// # Errors
	//
	//   - The readFile parameter returned an error.
	ReadFilesystemLayer(ctx context.Context, layer FilesystemLayer, readFile ReadFile) error
}

// ReadFile is a callback for each selected file in the FilesystemLayer. This
// is only called on regular files, which means it doesn't support tracking the
// directory that contains them. As this is usually backed by a tar file, it is
// possible the same name will be encountered more than once. It is also
// possible files are filtered out.
//
// # Parameters
//
// The parameters correspond with tar.Header fields and are unaltered when this
// is backed by a tar. The reader argument optionally reads from the current
// file until io.EOF. Use the size argument to be more precise.
type ReadFile func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error

// Image represents filesystem layers that make up an image on a specific
// Platform, parsed from the OCI manifest and
// configuration.
//
// See https://github.com/opencontainers/image-spec/blob/master/manifest.md
// and https://github.com/opencontainers/image-spec/blob/master/config.md
type Image interface {
	internal.CarOnly

	// Platform is the potentially empty platform. When present, this is
	// typically 'runtime.GOOS/runtime.GOARCH'. e.g. "darwin/amd64"
	Platform() string

	// FilesystemLayerCount is the count of layers, used to loop.
	FilesystemLayerCount() int

	// FilesystemLayer returns a FilesystemLayer given its index or nil if invalid.
	FilesystemLayer(int) FilesystemLayer

	fmt.Stringer
}

// FilesystemLayer is a reference to a non-empty, possibly zipped layer.
//
// See https://github.com/opencontainers/image-spec/blob/master/layer.md
type FilesystemLayer interface {
	internal.CarOnly

	// MediaType is the content type of this layer.
	//
	// # Examples
	//
	//   - MediaTypeOCIImageLayer
	//   - MediaTypeWasmImageLayer
	MediaType() string

	// Size is the size of the layer. For example, if it is a tar+gzip, this is
	// the compressed size in bytes of this "tar.gz"
	//
	// Note: When not a container image, or a shell command, the layer may have
	// Size zero.
	Size() int64

	// CreatedBy when present is the (usually Dockerfile) command that created
	// the layer
	CreatedBy() string

	// FileName is present when not a tar.
	FileName() string

	fmt.Stringer
}
