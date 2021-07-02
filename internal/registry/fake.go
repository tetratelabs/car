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

package registry

import (
	"context"
	"fmt"
	pathutil "path"
	"runtime"
	"time"

	"github.com/tetratelabs/car/internal"
)

type fakeRegistry struct {
	baseURL   string
	fakeLayer *internal.FilesystemLayer
}

// NewFake implements internal.NewRegistry for a fake registry
func NewFake(ctx context.Context, host, path string) internal.Registry {
	baseURL := fmt.Sprintf("mem://%s/v2/%s", host, path)
	fakeLayer := internal.FilesystemLayer{
		URL:       fmt.Sprintf("%s/blobs/%s", baseURL, "sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f"),
		MediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
		Size:      18500000,
	}
	return &fakeRegistry{baseURL: baseURL, fakeLayer: &fakeLayer}
}

func (m *fakeRegistry) String() string {
	return m.baseURL
}

func (m *fakeRegistry) GetImage(_ context.Context, tag, platform string) (*internal.Image, error) {
	if platform != pathutil.Join(runtime.GOOS, runtime.GOARCH) {
		return nil, fmt.Errorf("platform %s not found", platform)
	}
	return &internal.Image{
		URL:              fmt.Sprintf("mem://%s/%s", tag, platform),
		Platform:         platform,
		FilesystemLayers: []*internal.FilesystemLayer{m.fakeLayer},
	}, nil
}

func (m *fakeRegistry) ReadFilesystemLayer(_ context.Context, layer *internal.FilesystemLayer, readFile internal.ReadFile) error {
	if layer != m.fakeLayer {
		return fmt.Errorf("layer %v not found", layer)
	}
	modTime, err := time.Parse(time.RFC3339, "2021-06-21T15:05:05Z")
	if err != nil {
		return err
	}
	return readFile("usr/local/bin/envoy", 124804608, 0755, modTime, nil) // nolint:gocritic
}
