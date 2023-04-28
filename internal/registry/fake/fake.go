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

	"github.com/tetratelabs/car/internal"
)

type fakeRegistry struct {
	baseURL          string
	platform, tag    string
	filesystemLayers []*internal.FilesystemLayer
}

// NewRegistry implements internal.NewRegistry for a fake registry
func NewRegistry(_ context.Context, host string) (internal.Registry, error) {
	baseURL := fmt.Sprintf("fake://%s/v2", host)
	return &fakeRegistry{
		baseURL:          baseURL,
		platform:         "linux/amd64",
		tag:              "v1.0",
		filesystemLayers: fakeFilesystemLayers(baseURL),
	}, nil
}

func (f *fakeRegistry) String() string {
	return f.baseURL
}

func (f *fakeRegistry) GetImage(_ context.Context, path, tag, platform string) (*internal.Image, error) {
	if platform != "" && platform != f.platform {
		return nil, fmt.Errorf("platform %s not found", platform)
	}
	if tag != f.tag {
		return nil, fmt.Errorf("tag %s not found", tag)
	}
	return &internal.Image{
		URL:              fmt.Sprintf("%s/%s/manifests/%s", f.baseURL, path, tag),
		Platform:         f.platform,
		FilesystemLayers: f.filesystemLayers,
	}, nil
}

func (f *fakeRegistry) ReadFilesystemLayer(_ context.Context, layer *internal.FilesystemLayer, readFile internal.ReadFile) error {
	var files []*fakeFile
	for i, l := range f.filesystemLayers {
		if layer == l {
			files = fakeFiles[i]
			break
		}
	}
	if files == nil {
		return fmt.Errorf("layer %s not found", layer.URL)
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
func fakeFilesystemLayers(baseURL string) []*internal.FilesystemLayer {
	return []*internal.FilesystemLayer{
		{
			URL:       fmt.Sprintf("%s/blobs/%s", baseURL, "sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f"),
			MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
			Size:      30,
			CreatedBy: `/bin/sh -c #(nop) ADD file:d7fa3c26651f9204a5629287a1a9a6e7dc6a0bc6eb499e82c433c0c8f67ff46b in /`,
		},
		{
			URL:       fmt.Sprintf("%s/blobs/%s", baseURL, "sha256:15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2"),
			MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
			Size:      30,
			CreatedBy: `ADD build/* /usr/local/bin/ # buildkit`,
		},
		{
			URL:       fmt.Sprintf("%s/blobs/%s", baseURL, "sha256:1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81"),
			MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
			Size:      40,
			CreatedBy: `cmd /S /C powershell iex(iwr -useb https://moretrucks.io/install.ps1)`,
		},
		{
			URL:       fmt.Sprintf("%s/blobs/%s", baseURL, "sha256:6d2d8da2960b0044c22730be087e6d7b197ab215d78f9090a3dff8cb7c40c241"),
			MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
			Size:      50,
			CreatedBy: `ADD build/* /usr/local/sbin/ # buildkit`,
		},
	}
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
