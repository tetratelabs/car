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
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	pathutil "path"
	"strings"

	"github.com/tetratelabs/car/internal"
	"github.com/tetratelabs/car/internal/httpclient"
	"github.com/tetratelabs/car/internal/registry/docker"
	"github.com/tetratelabs/car/internal/registry/github"
)

type registry struct {
	baseURL    string
	httpClient httpclient.HTTPClient
}

// New returns a new instance of a remote registry
// * host is the registry host.
//   * Empty ("") implies the path is a DockerHub image like "alpine" or "envoyproxy/envoy".
// * path is the image path which must include at least one slash, possibly more than two.
//   * The only paths allowed to exclude a slash are DockerHub official images like "alpine"
func New(ctx context.Context, host, path string) internal.Registry {
	if host == "" || host == "docker.io" {
		host = "index.docker.io"
	}
	if !strings.Contains(path, "/") {
		path = pathutil.Join("library", path)
	}
	transport := httpClientTransport(ctx, host, path)
	baseURL := fmt.Sprintf("https://%s/v2/%s", host, path)
	return &registry{baseURL: baseURL, httpClient: httpclient.New(transport)}
}

// httpClientTransport returns the http.Client Transport appropriate for the registry
func httpClientTransport(ctx context.Context, host, path string) http.RoundTripper {
	switch host {
	case "index.docker.io":
		return docker.NewRoundTripper(path)
	case "ghcr.io":
		return github.NewRoundTripper()
	default:
		return httpclient.TransportFromContext(ctx)
	}
}

func (r *registry) String() string {
	return r.baseURL
}

func (r *registry) GetImage(ctx context.Context, tag, platform string) (*internal.Image, error) {
	// A tag can respond with either a multi-platform image or a single one, so we have to handle either.
	images, err := r.getImageManifests(ctx, tag, platform)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("image tag %s not found", tag)
	}

	// History (created_by for each layer) is not in the manifest, rather the config JSON.
	configs, err := r.getImageConfigs(ctx, images)
	if err != nil {
		return nil, err
	}

	// Combine the two sources into the Image internal we need.
	var result []*internal.Image
	lastOSVersion := ""
	for i := range images {
		p := pathutil.Join(configs[i].OS, configs[i].Architecture)
		if platform == p && configs[i].OSVersion >= lastOSVersion {
			lastOSVersion = configs[i].OSVersion
			result = append(result, newImage(r.baseURL, images[i], configs[i]))
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("image tag %s not found for platform %s", tag, platform)
	}
	return result[len(result)-1], nil
}

func (r *registry) getImageManifests(ctx context.Context, tag, platform string) ([]*imageManifestV1, error) {
	header := http.Header{}
	header.Add("Accept", acceptImageIndexV1)
	header.Add("Accept", acceptImageManifestV1)

	url := fmt.Sprintf("%s/manifests/%s", r.baseURL, tag)
	body, mediaType, err := r.httpClient.Get(ctx, url, header)
	if err != nil {
		return nil, err
	}
	defer body.Close()         //nolint
	b, err := io.ReadAll(body) // fully read the response
	if err != nil {
		return nil, err
	}

	switch {
	case strings.Contains(acceptImageIndexV1, mediaType):
		index := imageIndexV1{}
		if err = json.Unmarshal(b, &index); err != nil {
			return nil, fmt.Errorf("error unmarshalling image index from %s: %w", url, err)
		}
		return r.getMultiPlatformManifests(ctx, &index, platform)
	case strings.Contains(acceptImageManifestV1, mediaType):
		manifest := imageManifestV1{}
		if err = json.Unmarshal(b, &manifest); err != nil {
			return nil, fmt.Errorf("error unmarshalling image manifest from %s: %w", url, err)
		}
		manifest.URL = url
		return []*imageManifestV1{&manifest}, nil
	default:
		return nil, fmt.Errorf("unknown mediaType %s from %s: %w", mediaType, url, err)
	}
}

func (r *registry) getMultiPlatformManifests(ctx context.Context, index *imageIndexV1, platform string) ([]*imageManifestV1, error) {
	var manifests []*imageManifestV1 //nolint:prealloc
	for _, ref := range index.Manifests {
		p := pathutil.Join(ref.Platform.OS, ref.Platform.Architecture)
		if p != platform {
			continue
		}
		url := fmt.Sprintf("%s/manifests/%s", r.baseURL, ref.Digest)
		manifest := imageManifestV1{}
		if err := r.httpClient.GetJSON(ctx, url, ref.MediaType, &manifest); err != nil {
			return nil, fmt.Errorf("error getting image ref for platform %v: %w", ref.Platform, err)
		}
		manifest.URL = url
		manifests = append(manifests, &manifest)
	}
	return manifests, nil
}

func (r *registry) getImageConfigs(ctx context.Context, images []*imageManifestV1) ([]*imageConfigV1, error) {
	var configs = make([]*imageConfigV1, len(images))
	for i, image := range images {
		if !strings.Contains(acceptImageConfigV1, image.Config.MediaType) {
			return nil, fmt.Errorf("invalid config media type in image %v", image)
		}
		url := fmt.Sprintf("%s/blobs/%s", r.baseURL, image.Config.Digest)
		config := imageConfigV1{}
		err := r.httpClient.GetJSON(ctx, url, image.Config.MediaType, &config)
		if err != nil {
			return nil, fmt.Errorf("error getting image config from %s: %w", url, err)
		}
		configs[i] = &config
	}
	return configs, nil
}

func (r *registry) ReadFilesystemLayer(ctx context.Context, layer *internal.FilesystemLayer, readFile internal.ReadFile) error {
	header := http.Header{}
	header.Add("Accept", layer.MediaType)
	body, _, err := r.httpClient.Get(ctx, layer.URL, header)
	if err != nil {
		return err
	}
	defer body.Close() //nolint
	zSrc, err := gzip.NewReader(body)
	if err != nil {
		return err
	}
	defer zSrc.Close() //nolint

	tr := tar.NewReader(zSrc)
	for {
		th, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Skip directories, symbolic links, block devices, etc.
		if th.Typeflag != tar.TypeReg {
			continue
		}

		// We currently don't implement deleting files from the list
		// https://github.com/opencontainers/image-spec/blob/859973e32ccae7b7fc76b40b762c9fff6e912f9e/layer.md#whiteouts
		if strings.Contains(th.Name, ".wh.") {
			continue
		}
		if err := readFile(th.Name, th.Size, th.Mode, th.ModTime, tr); err != nil {
			return fmt.Errorf("error calling readFile on %s: %w", th.Name, err)
		}
	}
	return nil
}
