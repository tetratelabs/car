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

package fake

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/tetratelabs/car/api"
	"github.com/tetratelabs/car/internal"
)

// image implements api.Image
type image struct {
	internal.CarOnly

	platform string
}

// Platform implements the same method as documented on api.Image
func (i image) Platform() string {
	return i.platform
}

// FilesystemLayerCount implements the same method as documented on api.Image
func (i image) FilesystemLayerCount() int {
	return len(fakeFilesystemLayers)
}

// FilesystemLayer implements the same method as documented on api.Image
func (i image) FilesystemLayer(idx int) api.FilesystemLayer {
	if idx < 0 || idx >= i.FilesystemLayerCount() {
		return nil
	}
	return fakeFilesystemLayers[idx]
}

// String implements fmt.Stringer
func (i image) String() string {
	return i.platform
}

// filesystemLayer is a reference to a non-empty, possibly zipped layer.
//
// See https://github.com/opencontainers/image-spec/blob/master/layer.md
type filesystemLayer struct {
	internal.CarOnly

	sha256    string
	mediaType string
	size      int64
	createdBy string
	fileName  string
}

// MediaType implements the same method as documented on api.FilesystemLayer
func (f filesystemLayer) MediaType() string {
	return f.mediaType
}

// Size implements the same method as documented on api.FilesystemLayer
func (f filesystemLayer) Size() int64 {
	return f.size
}

// CreatedBy implements the same method as documented on api.FilesystemLayer
func (f filesystemLayer) CreatedBy() string {
	return f.createdBy
}

// FileName implements the same method as documented on api.FilesystemLayer
func (f filesystemLayer) FileName() string {
	return f.fileName
}

// String implements fmt.Stringer
func (f filesystemLayer) String() string {
	return f.sha256
}

type fakeRegistry struct {
	internal.CarOnly

	host          string
	platform, tag string
}

var Registry = &fakeRegistry{
	platform: "linux/amd64",
	tag:      "v1.0",
}

func (f *fakeRegistry) GetImage(_ context.Context, ref api.Reference, platform string) (api.Image, error) {
	if platform != "" && platform != f.platform {
		return nil, fmt.Errorf("platform %s not found", platform)
	}
	if ref.Tag() != f.tag {
		return nil, fmt.Errorf("tag %s not found", ref.Tag())
	}
	return image{platform: f.platform}, nil
}

func (f *fakeRegistry) ReadFilesystemLayer(_ context.Context, layer api.FilesystemLayer, readFile api.ReadFile) error {
	sha256 := layer.(filesystemLayer).sha256
	var files []*fakeFile
	for i := range fakeFilesystemLayers {
		if sha256 == fakeFilesystemLayers[i].sha256 {
			files = fakeFiles[i]
			break
		}
	}
	if files == nil {
		return fmt.Errorf("layer %s not found", sha256)
	}
	for i, file := range files {
		modTime, err := time.Parse(time.RFC3339, file.modTimeRFC3339)
		if err != nil {
			return err
		}

		// make a fake file with contents that differ based on the index (this is to tell apart in debugger)
		fakeFile := make([]byte, file.size)
		for j := 0; j < len(fakeFile); j++ {
			fakeFile[j] = byte(i)
		}

		err = readFile(file.name, file.size, file.mode, modTime, bytes.NewReader(fakeFile))
		if err != nil {
			return err
		}
	}
	return nil
}

// fakeFilesystemLayers is pair-indexed with fakeFiles
var fakeFilesystemLayers = []filesystemLayer{
	{
		sha256:    "4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f",
		mediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
		size:      30,
		createdBy: `/bin/sh -c #(nop) ADD file:d7fa3c26651f9204a5629287a1a9a6e7dc6a0bc6eb499e82c433c0c8f67ff46b in /`,
	},
	{
		sha256:    "15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2",
		mediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
		size:      30,
		createdBy: `ADD build/* /usr/local/bin/ # buildkit`,
	},
	{
		sha256:    "1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81",
		mediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
		size:      40,
		createdBy: `cmd /S /C powershell iex(iwr -useb https://moretrucks.io/install.ps1)`,
	},
	{
		sha256:    "6d2d8da2960b0044c22730be087e6d7b197ab215d78f9090a3dff8cb7c40c241",
		mediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
		size:      50,
		createdBy: `ADD build/* /usr/local/sbin/ # buildkit`,
	},
}

type fakeFile struct {
	name           string
	size           int64
	mode           os.FileMode
	modTimeRFC3339 string
}

// fakeFiles is pair-indexed with fakeFilesystemLayers.
// The fake data intentionally overlaps on "usr/local" for testing. Even if weird, it adds windows paths.
var fakeFiles = [][]*fakeFile{
	{
		{"bin/apple.txt", 10, 0o640 & os.ModePerm, "2020-06-07T06:28:15Z"},
		{"usr/local/bin/boat", 20, 0o755 & os.ModePerm, "2021-04-16T22:53:09Z"},
	},
	{
		{"usr/local/bin/car", 30, 0o755 & os.ModePerm, "2021-05-12T03:53:29Z"},
	},
	{
		{"Files/ProgramData/truck/bin/truck.exe", 40, 0o644 & os.ModePerm, "2021-05-12T03:53:15Z"},
	},
	{
		{"usr/local/sbin/car", 50, 0o755 & os.ModePerm, "2021-05-12T03:53:29Z"},
	},
}
