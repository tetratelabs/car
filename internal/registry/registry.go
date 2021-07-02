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
	"sort"
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

// New implements internal.NewRegistry for a remote registry
func New(ctx context.Context, host, path string) internal.Registry {
	if host == "" || host == "docker.io" {
		host = "index.docker.io"
	}
	if !strings.Contains(path, "/") {
		path = pathutil.Join("library", path)
	}
	transport := httpClientTransport(ctx, host, path)
	scheme := "https"
	if strings.HasSuffix(host, ":5000") { // well-known plain text port. ex `docker run registry:2`
		scheme = "http"
	}
	baseURL := fmt.Sprintf("%s://%s/v2/%s", scheme, host, path)
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
	image, err := r.getImageManifest(ctx, tag, platform)
	if err != nil {
		return nil, err
	}

	// History (created_by for each layer) is not in the manifest, rather the config JSON.
	config, err := r.getImageConfig(ctx, image)
	if err != nil {
		return nil, err
	}

	// In a single-platform image, we won't know the platform until we got the config. Double-check!
	platforms := map[string]string{}
	if p := pathutil.Join(config.OS, config.Architecture); p != "" {
		platforms[p] = ""
	}
	if _, err = requireValidPlatform(platform, platforms); err != nil {
		return nil, err
	}

	// Combine the two sources into the Image we need.
	return newImage(r.baseURL, image, config), nil
}

func (r *registry) getImageManifest(ctx context.Context, tag, platform string) (*imageManifestV1, error) {
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
		return r.findPlatformManifest(ctx, &index, platform)
	case strings.Contains(acceptImageManifestV1, mediaType):
		manifest := imageManifestV1{}
		if err = json.Unmarshal(b, &manifest); err != nil {
			return nil, fmt.Errorf("error unmarshalling image manifest from %s: %w", url, err)
		}
		manifest.URL = url
		return &manifest, nil
	default:
		return nil, fmt.Errorf("unknown mediaType %s from %s: %w", mediaType, url, err)
	}
}

func (r *registry) findPlatformManifest(ctx context.Context, index *imageIndexV1, platform string) (*imageManifestV1, error) {
	platformToURL := map[string]string{} // duplicate keys are possible with os.version
	platformToOSVersion := map[string]string{}
	urlToMediaType := map[string]string{}

	for _, ref := range index.Manifests {
		p := pathutil.Join(ref.Platform.OS, ref.Platform.Architecture)
		if p == "" {
			continue // skip unknown platform
		}
		url := fmt.Sprintf("%s/manifests/%s", r.baseURL, ref.Digest)
		lastOSVersion := platformToOSVersion[p]
		if ref.Platform.OSVersion >= lastOSVersion {
			platformToURL[p] = url
			urlToMediaType[url] = ref.MediaType
			platformToOSVersion[p] = ref.Platform.OSVersion
		}
	}

	var err error
	if platform, err = requireValidPlatform(platform, platformToURL); err != nil {
		return nil, err
	}

	url := platformToURL[platform]
	mediaType := urlToMediaType[url]

	manifest := imageManifestV1{}
	if err := r.httpClient.GetJSON(ctx, url, mediaType, &manifest); err != nil {
		return nil, fmt.Errorf("error getting image ref for platform %s: %w", platform, err)
	}
	manifest.URL = url
	return &manifest, nil
}

func requireValidPlatform(platform string, platforms map[string]string) (string, error) {
	// While possible to pull a manifest with no platform information, we currently error as it could
	// be a sign of a bug in the JSON. We can change this to be allowed if platform == "" as needed.
	if len(platforms) == 0 {
		return "", fmt.Errorf("image config contains no platform information")
	}

	// If we are platform-agnostic return the only platform or error if it is ambiguous
	if platform == "" {
		if len(platforms) == 1 {
			for p := range platforms {
				return p, nil
			}
		}
		return "", fmt.Errorf("choose a platform: %s", sortedKeyString(platforms))
	}

	// see if the desired platform is present. Otherwise
	if _, ok := platforms[platform]; ok {
		return platform, nil
	}
	return "", fmt.Errorf("%s is not a supported platform: %s", platform, sortedKeyString(platforms))
}

func sortedKeyString(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return strings.Join(keys, ", ")
}

func (r *registry) getImageConfig(ctx context.Context, image *imageManifestV1) (*imageConfigV1, error) {
	if !strings.Contains(acceptImageConfigV1, image.Config.MediaType) {
		return nil, fmt.Errorf("invalid config media type in image %v", image)
	}
	url := fmt.Sprintf("%s/blobs/%s", r.baseURL, image.Config.Digest)
	config := imageConfigV1{}
	if err := r.httpClient.GetJSON(ctx, url, image.Config.MediaType, &config); err != nil {
		return nil, fmt.Errorf("error getting image config from %s: %w", url, err)
	}
	return &config, nil
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
