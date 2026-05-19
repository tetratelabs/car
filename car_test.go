// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package car_test

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car"
	"github.com/tetratelabs/car/api"
	"github.com/tetratelabs/car/internal/httpclient"
	"github.com/tetratelabs/car/internal/test/httptest"
)

//go:embed internal/registry/testdata/json/envoy-v1.38.0-linux-amd64-vnd.oci.image.manifest.v1.json
var linuxAmd64ManifestJSON []byte

//go:embed internal/registry/testdata/json/envoy-v1.38.0-linux-amd64-vnd.oci.image.config.v1.json
var linuxAmd64ConfigJSON []byte

func TestNewRegistry_DockerHub(t *testing.T) {
	ts := httptest.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Host {
		case "auth.docker.io":
			assert.Equal(t, "/token", r.URL.Path)
			assert.Contains(t, r.URL.RawQuery, "scope=repository:envoyproxy/envoy:pull")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(struct {
				Token string `json:"token"`
			}{"test-token"})
		case "index.docker.io":
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			switch {
			case strings.Contains(r.URL.Path, "/manifests/"):
				w.Header().Set("Content-Type", api.MediaTypeOCIImageManifest)
				w.Write(linuxAmd64ManifestJSON)
			case strings.Contains(r.URL.Path, "/blobs/"):
				w.Header().Set("Content-Type", api.MediaTypeOCIImageConfig)
				w.Write(linuxAmd64ConfigJSON)
			}
		default:
			http.Error(w, "unexpected host: "+r.Host, http.StatusInternalServerError)
		}
	}))

	ctx := httpclient.ContextWithTransport(t.Context(), ts.Client().Transport)

	ref, err := car.ParseReference("envoyproxy/envoy:v1.38.0")
	require.NoError(t, err)

	r, err := car.NewRegistry(ctx, ref.Domain())
	require.NoError(t, err)

	img, err := r.GetImage(ctx, ref, "linux/amd64")
	require.NoError(t, err)
	require.Equal(t, "linux/amd64", img.Platform())
	require.Equal(t, 9, img.FilesystemLayerCount())
}
