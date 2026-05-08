// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	_ "embed"
	"io"
	"io/fs"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/api"
	"github.com/tetratelabs/car/internal/httpclient"
	"github.com/tetratelabs/car/internal/reference"
	"github.com/tetratelabs/car/internal/registry/docker"
	"github.com/tetratelabs/car/internal/registry/github"
	"github.com/tetratelabs/car/internal/test/httptest"
)

type response struct {
	contentType string
	body        []byte
}

func TestNew(t *testing.T) {
	tests := []struct{ name, host, expectedBaseURL string }{
		{
			name:            "docker",
			host:            "index.docker.io",
			expectedBaseURL: "https://index.docker.io/v2",
		},
		{
			name:            "ghcr.io",
			host:            "ghcr.io",
			expectedBaseURL: "https://ghcr.io/v2",
		},
		{
			name:            "ghcr.io multiple slashes",
			host:            "ghcr.io",
			expectedBaseURL: "https://ghcr.io/v2",
		},
		{
			name:            "port 5443 is https",
			host:            "localhost:5443",
			expectedBaseURL: "https://localhost:5443/v2",
		},
		{
			name:            "port 5000 is plain text (localhost)",
			host:            "localhost:5000",
			expectedBaseURL: "http://localhost:5000/v2",
		},
		{
			name:            "port 5000 is plain text (127.0.0.1)",
			host:            "127.0.0.1:5000",
			expectedBaseURL: "http://127.0.0.1:5000/v2",
		},
		{
			name:            "port 5000 is plain text (e.g. docker compose)",
			host:            "registry:5000",
			expectedBaseURL: "http://registry:5000/v2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := New(t.Context(), tc.host)
			require.NoError(t, err)
			require.IsType(t, &registry{}, r)
			require.Equal(t, tc.expectedBaseURL, r.(*registry).baseURL)
			require.NotNil(t, r.(*registry).httpClient)
		})
	}
}

func TestHttpClientTransport(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		host     string
		expected http.RoundTripper
	}{
		{
			name:     "default nothing in context",
			ctx:      t.Context(),
			expected: http.DefaultTransport,
		},
		{
			name:     "default something in context",
			ctx:      httpclient.ContextWithTransport(t.Context(), github.NewRoundTripper()),
			expected: github.NewRoundTripper(),
		},
		{
			name:     "Docker",
			ctx:      t.Context(),
			host:     "index.docker.io",
			expected: docker.NewRoundTripper(),
		},
		{
			name:     "GitHub",
			ctx:      t.Context(),
			host:     "ghcr.io",
			expected: github.NewRoundTripper(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transport := httpClientTransport(tc.ctx, tc.host)
			require.IsType(t, tc.expected, transport)
		})
	}
}

func TestGetImage(t *testing.T) {
	tests := []struct {
		name, platform string
		expected       image
		expectedErr    string
		responses      map[string]response
	}{
		{
			name:     "no platform",
			expected: imageTrivy,
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageManifest,
					trivyVndOciImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a": {
					api.MediaTypeUnknownImageConfig,
					[]byte("{}"),
				},
			},
		},
		{
			name:        "no platform wrong choice",
			platform:    "windows/amd64",
			expected:    imageTrivy,
			expectedErr: "image config contains no platform information",
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageManifest,
					trivyVndOciImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a": {
					api.MediaTypeUnknownImageConfig,
					[]byte("{}"),
				},
			},
		},
		{
			name:     "single platform multiple layers",
			platform: "windows/amd64",
			expected: imageWindows,
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageManifest,
					windowsVndDockerImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:00378fa4979bfcc7d1f5d33bb8cebe526395021801f9e233f8909ffc25a6f630": {
					api.MediaTypeDockerContainerImage,
					windowsVndDockerImageConfigV1Json,
				},
			},
		},
		{
			name:     "single platform implicit choice",
			expected: imageWindows,
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageManifest,
					windowsVndDockerImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:00378fa4979bfcc7d1f5d33bb8cebe526395021801f9e233f8909ffc25a6f630": {
					api.MediaTypeDockerContainerImage,
					windowsVndDockerImageConfigV1Json,
				},
			},
		},
		{
			name:        "single platform wrong choice",
			platform:    "linux/amd64",
			expectedErr: "linux/amd64 is not a supported platform: windows/amd64",
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageManifest,
					windowsVndDockerImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:00378fa4979bfcc7d1f5d33bb8cebe526395021801f9e233f8909ffc25a6f630": {
					api.MediaTypeDockerContainerImage,
					windowsVndDockerImageConfigV1Json,
				},
			},
		},
		{
			name:     "single platform multiple os.version chooses latest",
			platform: "darwin/amd64",
			expected: imageHomebrew,
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageIndex,
					homebrewVndOciImageIndexV1Json,
				},
				"/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672": {
					api.MediaTypeOCIImageManifest,
					homebrew113VndOciImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:a7f8bac78026ae40545531454c2ef4df75ec3de1c60f1d6923142fe4e44daf8a": {
					api.MediaTypeDockerContainerImage,
					homebrew113VndOciImageConfigV1Json,
				},
			},
		},
		{
			name:     "implicit platform multiple os.version chooses latest",
			expected: imageHomebrew,
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageIndex,
					homebrewVndOciImageIndexV1Json,
				},
				"/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672": {
					api.MediaTypeOCIImageManifest,
					homebrew113VndOciImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:a7f8bac78026ae40545531454c2ef4df75ec3de1c60f1d6923142fe4e44daf8a": {
					api.MediaTypeDockerContainerImage,
					homebrew113VndOciImageConfigV1Json,
				},
			},
		},
		{
			name:     "index skips manifest missing platform",
			expected: imageHomebrew,
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageIndex,
					[]byte(`{
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:0da7ea4ca0f3615ace3b2223248e0baed539223df62d33d4c1a1e23346329057"
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672",
      "platform": {
        "architecture": "amd64",
        "os": "darwin",
        "os.version": "macOS 11.3"
      }
    }
  ]
}`),
				},
				"/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672": {
					api.MediaTypeOCIImageManifest,
					homebrew113VndOciImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:a7f8bac78026ae40545531454c2ef4df75ec3de1c60f1d6923142fe4e44daf8a": {
					api.MediaTypeDockerContainerImage,
					homebrew113VndOciImageConfigV1Json,
				},
			},
		},
		{
			name:        "single platform multiple os.version wrong choice",
			platform:    "windows/amd64",
			expectedErr: "windows/amd64 is not a supported platform: darwin/amd64",
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeOCIImageIndex,
					homebrewVndOciImageIndexV1Json,
				},
				"/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672": {
					api.MediaTypeOCIImageManifest,
					homebrew113VndOciImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:a7f8bac78026ae40545531454c2ef4df75ec3de1c60f1d6923142fe4e44daf8a": {
					api.MediaTypeDockerContainerImage,
					homebrew113VndOciImageConfigV1Json,
				},
			},
		},
		{
			name:     "chooses correct platform (linux/amd64)",
			platform: "linux/amd64",
			expected: imageLinuxAmd64,
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeDockerManifestList,
					linuxVndDockerImageIndexV1Json,
				},
				"/v2/user/repo/manifests/sha256:a85fb4ac4bce750803b9db6306152db456864486e36f61559731f9033a9293f0": {
					api.MediaTypeOCIImageManifest,
					linuxAmd64VndOciImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:ba952938f316ef8eb82a35a75bc79232a57233d20ce5ee14d63a6640d4257508": {
					api.MediaTypeOCIImageConfig,
					linuxAmd64VndOciImageConfigV1Json,
				},
			},
		},
		{
			name:     "multi-platform correct choice (linux/arm64)",
			platform: "linux/arm64",
			expected: imageLinuxArm64,
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeDockerManifestList,
					linuxVndDockerImageIndexV1Json,
				},
				"/v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c": {
					api.MediaTypeOCIImageManifest,
					linuxArm64VndDockerImageManifestV1Json,
				},
				"/v2/user/repo/blobs/sha256:a76857bf7e536baff5d0e4b316f1197dff0763bef3d9405f00e63f0deddb7447": {
					api.MediaTypeDockerContainerImage,
					linuxArm64VndDockerImageConfigV1Json,
				},
			},
		},
		{
			name:        "multi-platform, but no manifests",
			expectedErr: "image config contains no platform information",
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeDockerManifestList,
					[]byte(`{"manifests": []}`),
				},
			},
		},
		{
			name:        "multi-platform, all manifests have no platform",
			expectedErr: "image config contains no platform information",
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeDockerManifestList,
					[]byte(`{
  "manifests": [
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c",
      "size": 2403
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f",
      "size": 2403
    }
  ]
}`),
				},
			},
		},
		{
			name:        "multi-platform ambiguous",
			expectedErr: "choose a platform: linux/amd64, linux/arm64",
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeDockerManifestList,
					linuxVndDockerImageIndexV1Json,
				},
			},
		},
		{
			name:        "multi-platform wrong choice",
			platform:    "windows/arm64",
			expectedErr: "windows/arm64 is not a supported platform: linux/amd64, linux/arm64",
			responses: map[string]response{
				"/v2/user/repo/manifests/v1.0": {
					api.MediaTypeDockerManifestList,
					linuxVndDockerImageIndexV1Json,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp, ok := tc.responses[r.URL.Path]
				if !ok {
					http.Error(w, "unexpected request: "+r.URL.Path, http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", resp.contentType)
				w.Write(resp.body)
			}))
			ctx := httpclient.ContextWithTransport(t.Context(), ts.Client().Transport)

			ref := reference.MustParse("user/repo:v1.0")
			r, err := New(ctx, "test:5000")
			require.NoError(t, err)
			i, err := r.GetImage(ctx, ref, tc.platform)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.IsType(t, image{}, i)
				require.Equal(t, tc.expected.filesystemLayers, i.(image).filesystemLayers)
				require.Equal(t, tc.expected, i)
			}
		})
	}
}

//go:embed testdata/add.wasm
var addWasm []byte

//go:embed testdata/test.tar.gz
var tarGz []byte

func TestReadFilesystemLayer(t *testing.T) {
	tests := []struct {
		name      string
		layer     filesystemLayer
		expected  api.ReadFile
		expectErr string
		responses map[string]response
	}{
		{
			name: "tar.gz",
			layer: filesystemLayer{
				url:       "http://test:5000/v2/user/repo/blobs/sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb",
				mediaType: api.MediaTypeDockerImageLayer,
				size:      int64(len(tarGz)),
				createdBy: `COPY hello / # buildkit`,
			},
			responses: map[string]response{
				"/v2/user/repo/blobs/sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb": {
					api.MediaTypeDockerImageLayer,
					tarGz,
				},
			},
			expected: func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
				require.Equal(t, "./hello/README.txt", name)
				require.Equal(t, int64(6), size)
				require.Equal(t, fs.FileMode(0o644), mode)
				require.NotZero(t, modTime.Unix())

				b, err := io.ReadAll(reader)
				require.NoError(t, err)
				require.Equal(t, "hello\n", string(b))

				return nil
			},
		},
		{
			name: "wasm",
			layer: filesystemLayer{
				url:       "http://test:5000/v2/user/repo/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b",
				mediaType: api.MediaTypeModuleWasmImageLayer,
				size:      int64(len(addWasm)),
				fileName:  "add.wasm",
			},
			responses: map[string]response{
				"/v2/user/repo/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b": {
					api.MediaTypeModuleWasmImageLayer,
					addWasm,
				},
			},
			expected: func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
				require.Equal(t, "add.wasm", name)
				require.Equal(t, int64(len(addWasm)), size)
				require.Equal(t, fs.FileMode(0o644), mode)
				require.NotZero(t, modTime.Unix())

				b, err := io.ReadAll(reader)
				require.NoError(t, err)
				require.Equal(t, addWasm, b)

				return nil
			},
		},
		{
			name: "wasm missing name",
			layer: filesystemLayer{
				url:       imageTrivy.filesystemLayers[0].url,
				mediaType: imageTrivy.filesystemLayers[0].mediaType,
			},
			responses: map[string]response{
				"/v2/user/repo/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b": {
					api.MediaTypeModuleWasmImageLayer,
					addWasm,
				},
			},
			expected: func(_ string, _ int64, _ os.FileMode, _ time.Time, _ io.Reader) error {
				t.Fatal("unexpected to call file when missing name")
				return nil
			},
			expectErr: "missing filename",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp, ok := tc.responses[r.URL.Path]
				if !ok {
					http.Error(w, "unexpected request: "+r.URL.Path, http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", resp.contentType)
				w.Write(resp.body)
			}))
			ctx := httpclient.ContextWithTransport(t.Context(), ts.Client().Transport)

			r, err := New(ctx, "test:5000")
			require.NoError(t, err)
			err = r.ReadFilesystemLayer(ctx, &tc.layer, tc.expected)
			if tc.expectErr != "" {
				require.EqualError(t, err, tc.expectErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSortedKeyString(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{"empty", map[string]string{}, ""},
		{"only one", map[string]string{"foo": "bar"}, "foo"},
		{"sorted", map[string]string{"baz": "qux", "foo": "bar"}, "baz, foo"},
		{"unsorted", map[string]string{"foo": "bar", "baz": "qux"}, "baz, foo"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, sortedKeyString(tc.input))
		})
	}
}
