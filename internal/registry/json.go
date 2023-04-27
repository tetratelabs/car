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
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/tetratelabs/car/api"
)

const (
	// opencontainersImageTitle holds the filename when api.MediaTypeWasmImageConfig or api.MediaTypeWasmImageLayer.
	opencontainersImageTitle = "org.opencontainers.image.title"

	// acceptImageConfigV1 are media-types for imageConfigV1
	acceptImageConfigV1 = api.MediaTypeOCIImageConfig + "," + api.MediaTypeDockerContainerImage + "," + api.MediaTypeUnknownImageConfig

	// acceptImageIndexV1 are media-types for imageIndexV1, a.k.a. multi-platform image.
	acceptImageIndexV1 = api.MediaTypeOCIImageIndex + "," + api.MediaTypeDockerManifestList

	// acceptImageManifestV1 are media-types for imageManifestV1
	acceptImageManifestV1 = api.MediaTypeOCIImageManifest + "," + api.MediaTypeDockerManifest
)

// imageConfigV1 represents OCI Registry "/v2/${Repository}/blobs/${Digest}" responses for these media-types:
// * api.MediaTypeOCIImageConfig
// * api.MediaTypeDockerContainerImage
//
// We rely on index correlation between imageConfigV1.History and imageManifestV1.Layers because "rootfs/diff_ids"
// don't match.
// See https://github.com/opencontainers/image-spec/blob/master/schema/config-schema.json
type imageConfigV1 struct {
	Architecture string      `json:"architecture"`
	Config       configV1    `json:"config,omitempty"`
	OS           string      `json:"os"`
	OSVersion    string      `json:"os.version,omitempty"`
	History      []historyV1 `json:"history,omitempty"`
}

type configV1 struct {
	Env        []string `json:"Env"`
	Entrypoint []string `json:"Entrypoint"`
	Cmd        []string `json:"Cmd"`
}

type historyV1 struct {
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer,omitempty"`
}

// imageIndexV1 represents OCI Registry "/v2/${Repository}/manifests/${Tag}"
//
// See acceptImageIndexV1 for its media types.
// See https://github.com/opencontainers/image-spec/blob/master/schema/image-index-schema.json
type imageIndexV1 struct {
	Manifests []*imageManifestReferenceV1 `json:"manifests"`
}

type imageManifestReferenceV1 struct {
	MediaType string     `json:"mediaType"`
	Digest    string     `json:"digest"`
	Platform  platformV1 `json:"platform"`
}
type platformV1 struct { // redefined here because of the dotted "os.version" json field name.
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
	OSVersion    string `json:"os.version,omitempty"`
}

// imageManifestV1 represents OCI Registry "/v2/${Repository}/manifests/${Tag}" responses for these media-types:
//
// See acceptImageManifestV1 for its media types
// See https://github.com/opencontainers/image-spec/blob/master/schema/image-manifest-schema.json
type imageManifestV1 struct {
	URL    string         // not in the JSON
	Config descriptorV1   `json:"config"`
	Layers []descriptorV1 `json:"layers"`
}

// See https://github.com/opencontainers/image-spec/blob/master/descriptor.md
type descriptorV1 struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations"`
}

var (
	// ignoredDockerDirectives are Dockerfile directives that don't result in a tarball which could contain a binary.
	// This is used because some versions of Docker don't set `"empty_layer": true` in the config JSON.
	// We can't use an allow list because "RUN", "ADD" and "COPY" are not always in "created_by", most notably in the
	// canonical images made by https://github.com/docker-library/bashbrew
	ignoredDockerDirectives = []string{
		"ARG",
		"CMD",
		"ENTRYPOINT",
		"ENV",
		"EXPOSE",
		"HEALTHCHECK",
		"LABEL",
		"MAINTAINER",
		"ONBUILD",
		"SHELL",
		"STOPSIGNAL",
		"USER",
		"VOLUME",
		"WORKDIR",
	}
	// skipCreatedByPattern is currently Windows-specific, but might be needed on other images if we find cases of it.
	// There are intentionally multiple spaces allowed as there are examples of it, caused by Moby joining on space:
	// https://github.com/moby/moby/blob/7b9275c0da707b030e62c96b679a976f31f929d3/image/v1/imagev1.go#L32
	skipCreatedByPattern = regexp.MustCompile(".* +(?:" + strings.Join(ignoredDockerDirectives, "|") + ") .*")
)

func newImage(baseURL string, manifest *imageManifestV1, config *imageConfigV1) api.Image {
	layers := filterLayers(baseURL, manifest, config)
	return image{
		url:              manifest.URL,
		platform:         path.Join(config.OS, config.Architecture),
		env:              config.Config.Env,
		entrypoint:       config.Config.Entrypoint,
		cmd:              config.Config.Cmd,
		filesystemLayers: layers,
	}
}

func filterLayers(baseURL string, manifest *imageManifestV1, config *imageConfigV1) []filesystemLayer {
	history := config.History
	if len(history) == 0 { // history is optional, so back-fill if empty
		history = make([]historyV1, len(manifest.Layers))
	}

	// we may not have the layers for the entire history
	var layers []filesystemLayer
	for j, k := 0, 0; j < len(manifest.Layers); j++ {
		l := manifest.Layers[j]
		for history[k].EmptyLayer {
			k++ // skip layers explicitly empty by recent Docker
		}
		h := history[k]
		k++
		if l.MediaType == api.MediaTypeDockerImageForeignLayer {
			continue // skip foreign URLs
		}

		if skipCreatedByPattern.MatchString(h.CreatedBy) {
			continue
		}

		url := fmt.Sprintf("%s/blobs/%s", baseURL, l.Digest)
		layers = append(layers, filesystemLayer{
			url:       url,
			mediaType: l.MediaType,
			size:      l.Size,
			createdBy: h.CreatedBy,
			fileName:  l.Annotations[opencontainersImageTitle],
		})
	}
	return layers
}
