// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain arg copy of the License at
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
	"bytes"
	"context"
	_ "embed"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/api"
	"github.com/tetratelabs/car/internal/httpclient"
	"github.com/tetratelabs/car/internal/reference"
	"github.com/tetratelabs/car/internal/registry/docker"
	"github.com/tetratelabs/car/internal/registry/github"
)

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
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			r, err := New(ctx, tc.host)
			require.NoError(t, err)
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
			ctx:      context.Background(),
			expected: http.DefaultTransport,
		},
		{
			name:     "default something in context",
			ctx:      httpclient.ContextWithTransport(context.Background(), github.NewRoundTripper()),
			expected: github.NewRoundTripper(),
		},
		{
			name:     "Docker",
			ctx:      context.Background(),
			host:     "index.docker.io",
			expected: docker.NewRoundTripper(),
		},
		{
			name:     "GitHub",
			ctx:      context.Background(),
			host:     "ghcr.io",
			expected: github.NewRoundTripper(),
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			transport := httpClientTransport(tc.ctx, tc.host)
			require.IsType(t, tc.expected, transport)
		})
	}
}

var indexOrManifestRequest = `GET /v2/user/repo/manifests/v1.0 HTTP/1.1
Host: test
Accept: application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json
Accept: application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.v2+json

`

var homebrewRequests = []string{indexOrManifestRequest, `GET /v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672 HTTP/1.1
Host: test
Accept: application/vnd.oci.image.manifest.v1+json

`, `GET /v2/user/repo/blobs/sha256:a7f8bac78026ae40545531454c2ef4df75ec3de1c60f1d6923142fe4e44daf8a HTTP/1.1
Host: test
Accept: application/vnd.oci.image.config.v1+json

`}

var homebrewMediaTypes = []string{
	"application/vnd.oci.image.index.v1+json",
	api.MediaTypeOCIImageManifest,
	api.MediaTypeDockerContainerImage,
}

var homebrewResponseBodies = [][]byte{
	homebrewVndOciImageIndexV1Json,
	homebrew113VndOciImageManifestV1Json,
	homebrew113VndOciImageConfigV1Json,
}

var trivyRequests = []string{indexOrManifestRequest, `GET /v2/user/repo/blobs/sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a HTTP/1.1
Host: test
Accept: application/vnd.unknown.config.v1+json

`}

var trivyMediaTypes = []string{
	api.MediaTypeOCIImageManifest,
	api.MediaTypeUnknownImageConfig,
}

var trivyResponseBodies = [][]byte{
	trivyVndOciImageManifestV1Json,
	trivyVndOciUnknownConfigV1Json,
}

var windowsRequests = []string{indexOrManifestRequest, `GET /v2/user/repo/blobs/sha256:00378fa4979bfcc7d1f5d33bb8cebe526395021801f9e233f8909ffc25a6f630 HTTP/1.1
Host: test
Accept: application/vnd.docker.container.image.v1+json

`}

var windowsMediaTypes = []string{
	api.MediaTypeOCIImageManifest,
	api.MediaTypeDockerContainerImage,
}

var windowsResponseBodies = [][]byte{
	windowsVndDockerImageManifestV1Json,
	windowsVndDockerImageConfigV1Json,
}

func TestGetImage(t *testing.T) {
	tests := []struct {
		name, platform     string
		expected           image
		expectedErr        string
		expectedRequests   []string
		responseMediaTypes []string
		responseBodies     [][]byte
	}{
		{
			name:               "no platform",
			expected:           imageTrivy,
			expectedRequests:   trivyRequests,
			responseMediaTypes: trivyMediaTypes,
			responseBodies:     trivyResponseBodies,
		},
		{
			name:               "no platform wrong choice",
			platform:           "windows/amd64",
			expected:           imageTrivy,
			expectedRequests:   trivyRequests,
			responseMediaTypes: trivyMediaTypes,
			responseBodies:     trivyResponseBodies,
			expectedErr:        "image config contains no platform information",
		},
		{
			name:               "single platform multiple layers",
			platform:           "windows/amd64",
			expected:           imageWindows,
			expectedRequests:   windowsRequests,
			responseMediaTypes: windowsMediaTypes,
			responseBodies:     windowsResponseBodies,
		},
		{
			name:               "single platform implicit choice",
			expected:           imageWindows,
			expectedRequests:   windowsRequests,
			responseMediaTypes: windowsMediaTypes,
			responseBodies:     windowsResponseBodies,
		},
		{
			name:               "single platform wrong choice",
			platform:           "linux/amd64",
			expectedRequests:   windowsRequests,
			responseMediaTypes: windowsMediaTypes,
			responseBodies:     windowsResponseBodies,
			expectedErr:        "linux/amd64 is not a supported platform: windows/amd64",
		},
		{
			name:               "single platform multiple os.version chooses latest",
			platform:           "darwin/amd64",
			expected:           imageHomebrew,
			expectedRequests:   homebrewRequests,
			responseMediaTypes: homebrewMediaTypes,
			responseBodies:     homebrewResponseBodies,
		},
		{
			name:               "implicit platform multiple os.version chooses latest",
			expected:           imageHomebrew,
			expectedRequests:   homebrewRequests,
			responseMediaTypes: homebrewMediaTypes,
			responseBodies:     homebrewResponseBodies,
		},
		{
			name:               "index skips manifest missing platform",
			expected:           imageHomebrew,
			expectedRequests:   homebrewRequests,
			responseMediaTypes: homebrewMediaTypes,
			responseBodies: [][]byte{
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
				homebrew113VndOciImageManifestV1Json,
				homebrew113VndOciImageConfigV1Json,
			},
		},
		{
			name:               "single platform multiple os.version wrong choice",
			platform:           "windows/amd64",
			expectedRequests:   homebrewRequests,
			responseMediaTypes: homebrewMediaTypes,
			responseBodies:     homebrewResponseBodies,
			expectedErr:        "windows/amd64 is not a supported platform: darwin/amd64",
		},
		{
			name:     "chooses correct platform (linux/amd64)",
			platform: "linux/amd64",
			expected: imageLinuxAmd64,
			expectedRequests: []string{indexOrManifestRequest, `GET /v2/user/repo/manifests/sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f HTTP/1.1
Host: test
Accept: application/vnd.docker.distribution.manifest.v2+json

`, `GET /v2/user/repo/blobs/sha256:33655f17f09318801873b70f89c1596ce38f41f6c074e2343d26e9b425f939ec HTTP/1.1
Host: test
Accept: application/vnd.docker.container.image.v1+json

`},
			responseMediaTypes: []string{
				api.MediaTypeDockerManifestList,
				api.MediaTypeOCIImageManifest,
				api.MediaTypeDockerContainerImage,
			},
			responseBodies: [][]byte{
				linuxVndDockerImageIndexV1Json,
				linuxAmd64VndDockerImageManifestV1Json,
				linuxAmd64VndDockerImageConfigV1Json,
			},
		},
		{
			name:     "multi-platform correct choice (linux/arm64)",
			platform: "linux/arm64",
			expected: imageLinuxArm64,
			expectedRequests: []string{indexOrManifestRequest, `GET /v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c HTTP/1.1
Host: test
Accept: application/vnd.docker.distribution.manifest.v2+json

`, `GET /v2/user/repo/blobs/sha256:a76857bf7e536baff5d0e4b316f1197dff0763bef3d9405f00e63f0deddb7447 HTTP/1.1
Host: test
Accept: application/vnd.docker.container.image.v1+json

`},
			responseMediaTypes: []string{
				api.MediaTypeDockerManifestList,
				api.MediaTypeOCIImageManifest,
				api.MediaTypeDockerContainerImage,
			},
			responseBodies: [][]byte{
				linuxVndDockerImageIndexV1Json,
				linuxArm64VndDockerImageManifestV1Json,
				linuxArm64VndDockerImageConfigV1Json,
			},
		},
		{
			name:     "multi-platform correct choice (linux/arm64)",
			platform: "linux/arm64",
			expected: imageLinuxArm64,
			expectedRequests: []string{indexOrManifestRequest, `GET /v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c HTTP/1.1
Host: test
Accept: application/vnd.docker.distribution.manifest.v2+json

`, `GET /v2/user/repo/blobs/sha256:a76857bf7e536baff5d0e4b316f1197dff0763bef3d9405f00e63f0deddb7447 HTTP/1.1
Host: test
Accept: application/vnd.docker.container.image.v1+json

`},
			responseMediaTypes: []string{
				api.MediaTypeDockerManifestList,
				api.MediaTypeOCIImageManifest,
				api.MediaTypeDockerContainerImage,
			},
			responseBodies: [][]byte{
				linuxVndDockerImageIndexV1Json,
				linuxArm64VndDockerImageManifestV1Json,
				linuxArm64VndDockerImageConfigV1Json,
			},
		},
		{
			name:               "multi-platform, but no manifests",
			expectedRequests:   []string{indexOrManifestRequest},
			responseMediaTypes: []string{api.MediaTypeDockerManifestList},
			responseBodies:     [][]byte{[]byte(`{"manifests": []}`)},
			expectedErr:        "image config contains no platform information",
		},
		{
			name:               "multi-platform, all manifests have no platform",
			expectedRequests:   []string{indexOrManifestRequest},
			responseMediaTypes: []string{api.MediaTypeDockerManifestList},
			responseBodies: [][]byte{[]byte(`{
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
}`)},
			expectedErr: "image config contains no platform information",
		},
		{
			name:               "multi-platform ambiguous",
			expectedRequests:   []string{indexOrManifestRequest},
			responseMediaTypes: []string{api.MediaTypeDockerManifestList},
			responseBodies:     [][]byte{linuxVndDockerImageIndexV1Json},
			expectedErr:        "choose a platform: linux/amd64, linux/arm64",
		},
		{
			name:               "multi-platform wrong choice",
			platform:           "windows/arm64",
			expectedRequests:   []string{indexOrManifestRequest},
			responseMediaTypes: []string{api.MediaTypeDockerManifestList},
			responseBodies:     [][]byte{linuxVndDockerImageIndexV1Json},
			expectedErr:        "windows/arm64 is not a supported platform: linux/amd64, linux/arm64",
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := httpclient.ContextWithTransport(context.Background(), &mock{
				t:                  t,
				requests:           tc.expectedRequests,
				responseBodies:     tc.responseBodies,
				responseMediaTypes: tc.responseMediaTypes,
			})

			ref := reference.MustParse("user/repo:v1.0")
			r, err := New(ctx, "test")
			require.NoError(t, err)
			i, err := r.GetImage(ctx, ref, tc.platform)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
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
		name, platform     string
		layer              filesystemLayer
		expected           api.ReadFile
		expectedErr        string
		expectedRequests   []string
		responseMediaTypes []string
		responseBodies     [][]byte
	}{
		{
			name: "tar.gz",
			layer: filesystemLayer{
				url:       "https://test/v2/user/repo/blobs/sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb",
				mediaType: api.MediaTypeDockerImageLayer,
				size:      int64(len(tarGz)),
				createdBy: `COPY hello / # buildkit`,
			},
			expectedRequests: []string{`GET /v2/user/repo/blobs/sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb HTTP/1.1
Host: test
Accept: application/vnd.docker.image.rootfs.diff.tar.gzip

`},
			responseMediaTypes: []string{api.MediaTypeDockerImageLayer},
			responseBodies:     [][]byte{tarGz},
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
				url:       "https://test/v2/user/repo/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b",
				mediaType: api.MediaTypeWasmImageLayer,
				size:      int64(len(addWasm)),
				fileName:  "add.wasm",
			},
			expectedRequests: []string{`GET /v2/user/repo/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b HTTP/1.1
Host: test
Accept: application/vnd.module.wasm.content.layer.v1+wasm

`},
			responseMediaTypes: []string{api.MediaTypeWasmImageLayer},
			responseBodies:     [][]byte{addWasm},
			expected: func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
				require.Equal(t, "add.wasm", name)
				require.Equal(t, int64(len(addWasm)), size)
				require.Equal(t, fs.FileMode(0o644), mode)
				require.NotZero(t, modTime.Unix())

				// verify the fake body exists
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
			expectedRequests: []string{`GET /v2/user/repo/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b HTTP/1.1
Host: test
Accept: application/vnd.module.wasm.content.layer.v1+wasm

`},
			responseMediaTypes: []string{api.MediaTypeWasmImageLayer},
			responseBodies:     [][]byte{addWasm},
			expected: func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
				t.Fatal("unexpected to call file when missing name")
				return nil
			},
			expectedErr: "missing filename",
		},
		{
			name: "invalid media type",
			layer: filesystemLayer{
				url:       imageTrivy.filesystemLayers[0].url,
				mediaType: "application/json",
			},
			expected: func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
				t.Fatal("unexpected to call file when missing name")
				return nil
			},
			expectedErr: "unexpected media type: application/json",
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := httpclient.ContextWithTransport(context.Background(), &mock{
				t:                  t,
				requests:           tc.expectedRequests,
				responseBodies:     tc.responseBodies,
				responseMediaTypes: tc.responseMediaTypes,
			})

			r, err := New(ctx, "test")
			require.NoError(t, err)
			err = r.ReadFilesystemLayer(ctx, tc.layer, tc.expected)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type mock struct {
	t                  *testing.T
	i                  int
	requests           []string
	responseMediaTypes []string
	responseBodies     [][]byte
}

func (m *mock) RoundTrip(req *http.Request) (*http.Response, error) {
	raw := new(bytes.Buffer)
	req.Write(raw) //nolint
	require.Lessf(m.t, m.i, len(m.requests), "bug: not enough requests")
	require.Lessf(m.t, m.i, len(m.responseBodies), "bug: not enough responseBodies")
	require.Lessf(m.t, m.i, len(m.responseMediaTypes), "bug: not enough responseMediaTypes")

	require.Equal(m.t, m.requests[m.i], strings.ReplaceAll(raw.String(), "\r\n", "\n"))

	body := m.responseBodies[m.i]
	mediaType := m.responseMediaTypes[m.i]
	m.i++
	return &http.Response{
		Status: "200 OK", StatusCode: http.StatusOK,
		Header: http.Header{"Content-Type": []string{mediaType}}, Body: io.NopCloser(bytes.NewReader(body)),
	}, nil
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
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, sortedKeyString(tc.input))
		})
	}
}
