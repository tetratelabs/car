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
	"regexp"
	"strings"

	"github.com/tetratelabs/car/internal"
)

const (
	mediaTypeDockerLayer = "application/vnd.docker.image.rootfs.diff.tar.gzip"
	mediaTypeOCILayer    = "application/vnd.oci.image.layer.v1.tar+gzip"
)

// isMediaTypeImageLayerV1 returns true for "tar.gz" layer types referenced by imageManifestV1.LayerDigests.
// These are "Accept" headers for the OCI Registry "/v2/${Repository}/blobs/${Digest}" endpoint.
func isMediaTypeImageLayerV1(mediaType string) bool {
	return mediaType == mediaTypeOCILayer || mediaType == mediaTypeDockerLayer
}

// imageConfigV1 represents OCI Registry "/v2/${Repository}/blobs/${Digest}" responses for these media-types:
// * "application/vnd.oci.image.config.v1+json"
// * "application/vnd.docker.container.image.v1+json"
//
// We rely on index correlation between imageConfigV1.History and imageManifestV1.Layers because "rootfs/diff_ids"
// don't match.
// See https://github.com/opencontainers/image-spec/blob/master/schema/config-schema.json
type imageConfigV1 struct {
	Architecture string      `json:"architecture"`
	OS           string      `json:"os"`
	OSVersion    string      `json:"os.version,omitempty"`
	History      []historyV1 `json:"history,omitempty"`
}

type historyV1 struct {
	CreatedBy  string `json:"created_by"`
	EmptyLayer bool   `json:"empty_layer,omitempty"`
}

// imageIndexV1 represents OCI Registry "/v2/${Repository}/manifests/${Tag}" responses for these media-types:
// * "application/vnd.oci.image.index.v1+json"
// * "application/vnd.docker.distribution.manifest.list.v2+json"
//
// See https://github.com/opencontainers/image-spec/blob/master/schema/image-index-schema.json
type imageIndexV1 struct {
	Manifests []imageManifestReferenceV1 `json:"manifests"`
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

// imageManifestV1 represents responses matching isMediaTypeImageManifestV1

// imageManifestV1 represents OCI Registry "/v2/${Repository}/manifests/${Tag}" responses for these media-types:
// * "application/vnd.oci.image.manifest.v1+json"
// * "application/vnd.docker.distribution.manifest.v2+json"
//
// See https://github.com/opencontainers/image-spec/blob/master/schema/image-manifest-schema.json
type imageManifestV1 struct {
	URL    string         // not in the JSON
	Config descriptorV1   `json:"config"`
	Layers []descriptorV1 `json:"layers"`
}

// See https://github.com/opencontainers/image-spec/blob/master/descriptor.md
type descriptorV1 struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
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

func newImage(baseURL string, manifest *imageManifestV1, config *imageConfigV1) *internal.Image {
	history := config.History
	if len(history) == 0 { // history is optional, so back-fill if empty
		history = make([]historyV1, len(manifest.Layers))
	}

	var layers []*internal.FilesystemLayer

	// we may not have the layers for the entire history
	for j, k := 0, 0; j < len(manifest.Layers); j++ {
		l := manifest.Layers[j]
		for history[k].EmptyLayer {
			k++ // skip layers explicitly empty by recent Docker
		}
		h := history[k]
		k++
		if !isMediaTypeImageLayerV1(l.MediaType) {
			continue // skip unsupported media types
		}

		if skipCreatedByPattern.MatchString(h.CreatedBy) {
			continue
		}

		url := fmt.Sprintf("%s/blobs/%s", baseURL, l.Digest)
		layer := &internal.FilesystemLayer{URL: url, MediaType: l.MediaType, Size: l.Size, CreatedBy: h.CreatedBy}
		layers = append(layers, layer)
	}
	return &internal.Image{URL: manifest.URL, Platform: config.OS + "/" + config.Architecture, FilesystemLayers: layers}
}
