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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal"
	"github.com/tetratelabs/car/internal/httpclient"
	"github.com/tetratelabs/car/internal/registry/docker"
	"github.com/tetratelabs/car/internal/registry/github"
)

func TestNew(t *testing.T) {
	tests := []struct{ name, host, path, expectedBaseURL string }{
		{
			name:            "docker familiar",
			host:            "",
			path:            "envoyproxy/envoy",
			expectedBaseURL: "https://index.docker.io/v2/envoyproxy/envoy",
		},
		{
			name:            "docker fully qualified",
			host:            "docker.io",
			path:            "envoyproxy/envoy",
			expectedBaseURL: "https://index.docker.io/v2/envoyproxy/envoy",
		},
		{
			name:            "docker familiar official",
			host:            "",
			path:            "alpine",
			expectedBaseURL: "https://index.docker.io/v2/library/alpine",
		},
		{
			name:            "docker unfamiliar official",
			host:            "docker.io",
			path:            "library/alpine",
			expectedBaseURL: "https://index.docker.io/v2/library/alpine",
		},
		{
			name:            "ghcr.io",
			host:            "ghcr.io",
			path:            "tetratelabs/car",
			expectedBaseURL: "https://ghcr.io/v2/tetratelabs/car",
		},
		{
			name:            "ghcr.io multiple slashes",
			host:            "ghcr.io",
			path:            "homebrew/core/envoy",
			expectedBaseURL: "https://ghcr.io/v2/homebrew/core/envoy",
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			r := New(ctx, tc.host, tc.path).(*registry)
			require.Equal(t, tc.expectedBaseURL, r.baseURL)
			require.NotNil(t, r.httpClient)
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
			expected: docker.NewRoundTripper(""),
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
			transport := httpClientTransport(tc.ctx, tc.host, "")
			require.IsType(t, tc.expected, transport)
		})
	}
}

var homebrewRequests = []string{`GET /v2/user/repo/manifests/v1.0 HTTP/1.1
Host: test
Accept: application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json
Accept: application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.v2+json

`, `GET /v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672 HTTP/1.1
Host: test
Accept: application/vnd.oci.image.manifest.v1+json

`, `GET /v2/user/repo/blobs/sha256:a7f8bac78026ae40545531454c2ef4df75ec3de1c60f1d6923142fe4e44daf8a HTTP/1.1
Host: test
Accept: application/vnd.oci.image.config.v1+json

`}

var homebrewMediaTypes = []string{
	"application/vnd.oci.image.index.v1+json",
	mediaTypeOCIImageManifest,
	mediaTypeDockerContainerImage,
}

var homebrewResponseBodies = [][]byte{
	homebrewVndOciImageIndexV1Json,
	homebrew113VndOciImageManifestV1Json,
	homebrew113VndOciImageConfigV1Json,
}

var linuxIndexRequest = `GET /v2/user/repo/manifests/v1.0 HTTP/1.1
Host: test
Accept: application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json
Accept: application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.v2+json

`

var windowsRequests = []string{`GET /v2/user/repo/manifests/v1.0 HTTP/1.1
Host: test
Accept: application/vnd.oci.image.index.v1+json,application/vnd.docker.distribution.manifest.list.v2+json
Accept: application/vnd.oci.image.manifest.v1+json,application/vnd.docker.distribution.manifest.v2+json

`, `GET /v2/user/repo/blobs/sha256:00378fa4979bfcc7d1f5d33bb8cebe526395021801f9e233f8909ffc25a6f630 HTTP/1.1
Host: test
Accept: application/vnd.docker.container.image.v1+json

`}

var windowsMediaTypes = []string{
	mediaTypeOCIImageManifest,
	mediaTypeDockerContainerImage,
}

var windowsResponseBodies = [][]byte{
	windowsVndDockerImageManifestV1Json,
	windowsVndDockerImageConfigV1Json,
}

func TestGetImage(t *testing.T) {
	tests := []struct {
		name, platform     string
		expected           *internal.Image
		expectedErr        string
		expectedRequests   []string
		responseMediaTypes []string
		responseBodies     [][]byte
	}{
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
			expected:           imageWindows,
			platform:           "linux/amd64",
			expectedRequests:   windowsRequests,
			responseMediaTypes: windowsMediaTypes,
			responseBodies:     windowsResponseBodies,
			expectedErr:        "tag v1.0 is for platform windows/amd64, not linux/amd64",
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
			name:               "single platform multiple os.version wrong choice",
			platform:           "windows/amd64",
			expected:           imageHomebrew,
			expectedRequests:   homebrewRequests,
			responseMediaTypes: homebrewMediaTypes,
			responseBodies:     homebrewResponseBodies,
			expectedErr:        "tag v1.0 is for platform darwin/amd64, not windows/amd64",
		},
		{
			name:     "chooses correct platform (linux/amd64)",
			platform: "linux/amd64",
			expected: imageLinuxAmd64,
			expectedRequests: []string{linuxIndexRequest, `GET /v2/user/repo/manifests/sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f HTTP/1.1
Host: test
Accept: application/vnd.docker.distribution.manifest.v2+json

`, `GET /v2/user/repo/blobs/sha256:33655f17f09318801873b70f89c1596ce38f41f6c074e2343d26e9b425f939ec HTTP/1.1
Host: test
Accept: application/vnd.docker.container.image.v1+json

`},
			responseMediaTypes: []string{
				mediaTypeDockerManifestList,
				mediaTypeOCIImageManifest,
				mediaTypeDockerContainerImage,
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
			expectedRequests: []string{linuxIndexRequest, `GET /v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c HTTP/1.1
Host: test
Accept: application/vnd.docker.distribution.manifest.v2+json

`, `GET /v2/user/repo/blobs/sha256:a76857bf7e536baff5d0e4b316f1197dff0763bef3d9405f00e63f0deddb7447 HTTP/1.1
Host: test
Accept: application/vnd.docker.container.image.v1+json

`},
			responseMediaTypes: []string{
				mediaTypeDockerManifestList,
				mediaTypeOCIImageManifest,
				mediaTypeDockerContainerImage,
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
			expectedRequests: []string{linuxIndexRequest, `GET /v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c HTTP/1.1
Host: test
Accept: application/vnd.docker.distribution.manifest.v2+json

`, `GET /v2/user/repo/blobs/sha256:a76857bf7e536baff5d0e4b316f1197dff0763bef3d9405f00e63f0deddb7447 HTTP/1.1
Host: test
Accept: application/vnd.docker.container.image.v1+json

`},
			responseMediaTypes: []string{
				mediaTypeDockerManifestList,
				mediaTypeOCIImageManifest,
				mediaTypeDockerContainerImage,
			},
			responseBodies: [][]byte{
				linuxVndDockerImageIndexV1Json,
				linuxArm64VndDockerImageManifestV1Json,
				linuxArm64VndDockerImageConfigV1Json,
			},
		},
		{
			name:               "multi-platform ambiguous",
			expected:           imageLinuxArm64,
			expectedRequests:   []string{linuxIndexRequest},
			responseMediaTypes: []string{mediaTypeDockerManifestList},
			responseBodies:     [][]byte{linuxVndDockerImageIndexV1Json},
			expectedErr:        "tag v1.0 is for platforms [linux/amd64 linux/arm64]: pick one",
		},
		{
			name:               "multi-platform wrong choice",
			platform:           "windows/arm64",
			expected:           imageLinuxArm64,
			expectedRequests:   []string{linuxIndexRequest},
			responseMediaTypes: []string{mediaTypeDockerManifestList},
			responseBodies:     [][]byte{linuxVndDockerImageIndexV1Json},
			expectedErr:        "tag v1.0 is for platforms [linux/amd64 linux/arm64], not windows/arm64",
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

			r := New(ctx, "test", "user/repo").(*registry)
			image, err := r.GetImage(ctx, "v1.0", tc.platform)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, image)
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
	return &http.Response{Status: "200 OK", StatusCode: http.StatusOK,
		Header: http.Header{"Content-Type": []string{mediaType}}, Body: io.NopCloser(bytes.NewReader(body))}, nil
}
