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
	baseURL string
	image   *internal.Image
	tag     string
}

// NewRegistry implements internal.NewRegistry for a fake registry
func NewRegistry(_ context.Context, host, path string) internal.Registry {
	baseURL := fmt.Sprintf("fake://%s/v2/%s", host, path)
	tag := "v1.0"
	return &fakeRegistry{baseURL, &internal.Image{
		URL:              fmt.Sprintf("%s/manifests/%s", baseURL, tag),
		Platform:         internal.OSLinux + "/" + internal.ArchAmd64,
		FilesystemLayers: fakeFilesystemLayers(baseURL),
	}, tag}
}

func (f *fakeRegistry) String() string {
	return f.baseURL
}

func (f *fakeRegistry) GetImage(_ context.Context, tag, platform string) (*internal.Image, error) {
	if platform != "" && platform != f.image.Platform {
		return nil, fmt.Errorf("platform %s not found", platform)
	}
	if tag != f.tag {
		return nil, fmt.Errorf("tag %s not found", tag)
	}
	return f.image, nil
}

func (f *fakeRegistry) ReadFilesystemLayer(_ context.Context, layer *internal.FilesystemLayer, readFile internal.ReadFile) error {
	var files []*fakeFile
	for i, l := range f.image.FilesystemLayers {
		if layer == l {
			files = fakeFiles[i]
			break
		}
	}
	if files == nil {
		return fmt.Errorf("layer %s not found", layer.URL)
	}
	for _, file := range files {
		modTime, err := time.Parse(time.RFC3339, file.modTimeRFC3339)
		if err != nil {
			return err
		}
		err = readFile(file.name, file.size, file.mode, modTime, bytes.NewReader([]byte{}))
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
			Size:      26697009,
			CreatedBy: `/bin/sh -c #(nop) ADD file:d7fa3c26651f9204a5629287a1a9a6e7dc6a0bc6eb499e82c433c0c8f67ff46b in / `,
		},
		{
			URL:       fmt.Sprintf("%s/blobs/%s", baseURL, "sha256:15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2"),
			MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
			Size:      2000000,
			CreatedBy: `ADD build/* /usr/local/bin/ # buildkit`,
		},
		{
			URL:       fmt.Sprintf("%s/blobs/%s", baseURL, "sha256:1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81"),
			MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
			Size:      4000000,
			CreatedBy: `cmd /S /C powershell iex(iwr -useb https://moretrucks.io/install.ps1)`,
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
		{"bin/apple.txt", 10, 0640 & os.ModePerm, "2020-06-07T06:28:15Z"},
		{"usr/local/bin/boat", 20, 0755 & os.ModePerm, "2021-04-16T22:53:09Z"},
	},
	{
		{"usr/local/bin/car", 30, 0755 & os.ModePerm, "2021-05-12T03:53:29Z"},
	},
	{
		{"Files/ProgramData/truck/bin/truck.exe", 40, 0444 & os.ModePerm, "2021-05-12T03:53:15Z"},
	},
}
