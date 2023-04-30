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
	_ "embed" // We embed the json files to harden the test build
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/api"
)

//go:embed testdata/json/homebrew-11.3-vnd.oci.image.config.v1.json
var homebrew113VndOciImageConfigV1Json []byte

func TestImageConfigV1_Homebrew(t *testing.T) {
	var v imageConfigV1
	require.NoError(t, json.Unmarshal(homebrew113VndOciImageConfigV1Json, &v))

	require.Equal(t, imageConfigV1{
		Architecture: "amd64",
		OS:           "darwin",
		OSVersion:    "macOS 11.3",
	}, v)
}

//go:embed testdata/json/homebrew-vnd.oci.image.index.v1.json
var homebrewVndOciImageIndexV1Json []byte

func TestImageIndexV1_Homebrew(t *testing.T) {
	var v imageIndexV1
	require.NoError(t, json.Unmarshal(homebrewVndOciImageIndexV1Json, &v))

	require.Equal(t, imageIndexV1{
		Manifests: []*imageManifestReferenceV1{
			{
				MediaType: api.MediaTypeOCIImageManifest,
				Digest:    "sha256:0da7ea4ca0f3615ace3b2223248e0baed539223df62d33d4c1a1e23346329057",
				Platform:  platformV1{"amd64", "darwin", "macOS 10.15.7"},
			},
			{
				MediaType: api.MediaTypeOCIImageManifest,
				Digest:    "sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672",
				Platform:  platformV1{"amd64", "darwin", "macOS 11.3"},
			},
		},
	}, v)
}

//go:embed testdata/json/homebrew-11.3-vnd.oci.image.manifest.v1.json
var homebrew113VndOciImageManifestV1Json []byte

func TestImageManifestV1_Homebrew(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(homebrew113VndOciImageManifestV1Json, &v))

	require.Equal(t, imageManifestV1{
		Config: descriptorV1{
			MediaType: api.MediaTypeOCIImageConfig,
			Digest:    "sha256:a7f8bac78026ae40545531454c2ef4df75ec3de1c60f1d6923142fe4e44daf8a",
		},
		Layers: []descriptorV1{
			{
				MediaType: api.MediaTypeOCIImageLayer,
				Digest:    "sha256:d03fb86b48336c8d3c0f3711cfc3df3557f9fb33c966ceb1caecae1653935e90",
				Size:      29405739,
			},
		},
	}, v)
}

var imageHomebrew = image{
	url:      "https://test/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672",
	platform: "darwin/amd64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "https://test/v2/user/repo/blobs/sha256:d03fb86b48336c8d3c0f3711cfc3df3557f9fb33c966ceb1caecae1653935e90",
			mediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			size:      29405739,
		},
	},
}

func TestNewImage_Homebrew(t *testing.T) {
	var i imageManifestV1
	require.NoError(t, json.Unmarshal(homebrew113VndOciImageManifestV1Json, &i))
	var c imageConfigV1
	require.NoError(t, json.Unmarshal(homebrew113VndOciImageConfigV1Json, &c))
	i.URL = "https://test/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672"
	require.Equal(t, imageHomebrew, newImage("https://test/v2/user/repo", &i, &c))
}

//go:embed testdata/json/linux-amd64-vnd.docker.container.image.v1.json
var linuxAmd64VndDockerImageConfigV1Json []byte

//go:embed testdata/json/linux-arm64-vnd.docker.container.image.v1.json
var linuxArm64VndDockerImageConfigV1Json []byte

func TestImageConfigV1_LinuxAmd64(t *testing.T) {
	var v imageConfigV1
	require.NoError(t, json.Unmarshal(linuxAmd64VndDockerImageConfigV1Json, &v))

	require.Equal(t, imageConfigV1{
		Architecture: "amd64",
		OS:           "linux",
		OSVersion:    "",
		History: []historyV1{
			{`/bin/sh -c #(nop) ADD file:d7fa3c26651f9204a5629287a1a9a6e7dc6a0bc6eb499e82c433c0c8f67ff46b in / `, false},
			{`/bin/sh -c set -xe 		&& echo '#!/bin/sh' > /usr/sbin/policy-rc.d 	&& echo 'exit 101' >> /usr/sbin/policy-rc.d 	&& chmod +x /usr/sbin/policy-rc.d 		&& dpkg-divert --local --rename --add /sbin/initctl 	&& cp -a /usr/sbin/policy-rc.d /sbin/initctl 	&& sed -i 's/^exit.*/exit 0/' /sbin/initctl 		&& echo 'force-unsafe-io' > /etc/dpkg/dpkg.cfg.d/docker-apt-speedup 		&& echo 'DPkg::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' > /etc/apt/apt.conf.d/docker-clean 	&& echo 'APT::Update::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' >> /etc/apt/apt.conf.d/docker-clean 	&& echo 'Dir::Cache::pkgcache ""; Dir::Cache::srcpkgcache "";' >> /etc/apt/apt.conf.d/docker-clean 		&& echo 'Acquire::Languages "none";' > /etc/apt/apt.conf.d/docker-no-languages 		&& echo 'Acquire::GzipIndexes "true"; Acquire::CompressionTypes::Order:: "gz";' > /etc/apt/apt.conf.d/docker-gzip-indexes 		&& echo 'Apt::AutoRemove::SuggestsImportant "false";' > /etc/apt/apt.conf.d/docker-autoremove-suggests`, false},
			{`/bin/sh -c [ -z "$(apt-get indextargets)" ]`, true},
			{`/bin/sh -c mkdir -p /run/systemd && echo 'docker' > /run/systemd/container`, false},
			{`/bin/sh -c #(nop)  CMD ["/bin/bash"]`, true},
			{`ARG TARGETPLATFORM`, true},
			{`RUN |1 TARGETPLATFORM=linux/amd64 /bin/sh -c apt-get update && apt-get upgrade -y     && apt-get install --no-install-recommends -y ca-certificates     && apt-get autoremove -y && apt-get clean     && rm -rf /tmp/* /var/tmp/*     && rm -rf /var/lib/apt/lists/* # buildkit`, false},
			{`RUN |1 TARGETPLATFORM=linux/amd64 /bin/sh -c mkdir -p /etc/envoy # buildkit`, false},
			{`ARG ENVOY_BINARY_SUFFIX=_stripped`, true},
			{`ADD linux/amd64/build_release_stripped/* /usr/local/bin/ # buildkit`, false},
			{`ADD configs/envoyproxy_io_proxy.yaml /etc/envoy/envoy.yaml # buildkit`, false},
			{`ADD linux/amd64/build_release/su-exec /usr/local/bin/ # buildkit`, false},
			{`RUN |2 TARGETPLATFORM=linux/amd64 ENVOY_BINARY_SUFFIX=_stripped /bin/sh -c chown root:root /usr/local/bin/su-exec && adduser --group --system envoy # buildkit`, false},
			{`EXPOSE map[10000/tcp:{}]`, true},
			{`COPY ci/docker-entrypoint.sh / # buildkit`, false},
			{`ENTRYPOINT ["/docker-entrypoint.sh"]`, true},
			{`CMD ["envoy" "-c" "/etc/envoy/envoy.yaml"]`, true},
		},
	}, v)
}

//go:embed testdata/json/linux-vnd.docker.distribution.manifest.list.v2.json
var linuxVndDockerImageIndexV1Json []byte

func TestImageIndexV1_Linux(t *testing.T) {
	var v imageIndexV1
	require.NoError(t, json.Unmarshal(linuxVndDockerImageIndexV1Json, &v))

	require.Equal(t, imageIndexV1{
		Manifests: []*imageManifestReferenceV1{
			{
				MediaType: api.MediaTypeDockerManifest,
				Digest:    "sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c",
				Platform:  platformV1{"arm64", "linux", ""},
			},
			{
				MediaType: api.MediaTypeDockerManifest,
				Digest:    "sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f",
				Platform:  platformV1{"amd64", "linux", ""},
			},
		},
	}, v)
}

//go:embed testdata/json/linux-amd64-vnd.docker.distribution.manifest.v2.json
var linuxAmd64VndDockerImageManifestV1Json []byte

//go:embed testdata/json/linux-arm64-vnd.docker.distribution.manifest.v2.json
var linuxArm64VndDockerImageManifestV1Json []byte

func TestImageManifestV1_LinuxAmd64(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(linuxAmd64VndDockerImageManifestV1Json, &v))

	require.Equal(
		t,
		imageManifestV1{
			Config: descriptorV1{
				MediaType: api.MediaTypeDockerContainerImage,
				Digest:    "sha256:33655f17f09318801873b70f89c1596ce38f41f6c074e2343d26e9b425f939ec",
			},
			Layers: []descriptorV1{
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:01bf7da0a88c9e37ae418d17c0aeed0621524848d80ccb9e38c67e7ab8e11928",
					Size:      26697009,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:f3b4a5f15c7a0722b4f22e61b5387317eaf2602c27ffb2bceac9a25f19fbd156",
					Size:      852,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:57ffbe87baa135002dddb7a7460082c5d6a352186e1be9464c5f31db81378824",
					Size:      188,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:e2f93437f92e69c54acb27971690e8fe49ba75783cc2e6d5b0efbaa971d73fba",
					Size:      2922771,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:21cb341b2283d5501142f9e4f9d1b1941138ccc0777b8914b18f842b42d1571c",
					Size:      120,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2",
					Size:      21729278,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:3e05f50f195e6d16485c6a693092169b274d399d3cce98a87dd36c007a6911c3",
					Size:      749,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81",
					Size:      3500,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:2fb3fe4b571942f3d49d9c0ab84550cfa3843936278ce4e58dba28934efeff72",
					Size:      1467,
				},
				{
					MediaType: api.MediaTypeDockerImageLayer,
					Digest:    "sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb",
					Size:      490,
				},
			},
		},
		v,
	)
}

var imageLinuxAmd64 = image{
	url:      "https://test/v2/user/repo/manifests/sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f",
	platform: "linux/amd64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "https://test/v2/user/repo/blobs/sha256:01bf7da0a88c9e37ae418d17c0aeed0621524848d80ccb9e38c67e7ab8e11928",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      26697009,
			createdBy: `/bin/sh -c #(nop) ADD file:d7fa3c26651f9204a5629287a1a9a6e7dc6a0bc6eb499e82c433c0c8f67ff46b in / `,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:f3b4a5f15c7a0722b4f22e61b5387317eaf2602c27ffb2bceac9a25f19fbd156",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      852,
			createdBy: `/bin/sh -c set -xe 		&& echo '#!/bin/sh' > /usr/sbin/policy-rc.d 	&& echo 'exit 101' >> /usr/sbin/policy-rc.d 	&& chmod +x /usr/sbin/policy-rc.d 		&& dpkg-divert --local --rename --add /sbin/initctl 	&& cp -a /usr/sbin/policy-rc.d /sbin/initctl 	&& sed -i 's/^exit.*/exit 0/' /sbin/initctl 		&& echo 'force-unsafe-io' > /etc/dpkg/dpkg.cfg.d/docker-apt-speedup 		&& echo 'DPkg::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' > /etc/apt/apt.conf.d/docker-clean 	&& echo 'APT::Update::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' >> /etc/apt/apt.conf.d/docker-clean 	&& echo 'Dir::Cache::pkgcache ""; Dir::Cache::srcpkgcache "";' >> /etc/apt/apt.conf.d/docker-clean 		&& echo 'Acquire::Languages "none";' > /etc/apt/apt.conf.d/docker-no-languages 		&& echo 'Acquire::GzipIndexes "true"; Acquire::CompressionTypes::Order:: "gz";' > /etc/apt/apt.conf.d/docker-gzip-indexes 		&& echo 'Apt::AutoRemove::SuggestsImportant "false";' > /etc/apt/apt.conf.d/docker-autoremove-suggests`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:57ffbe87baa135002dddb7a7460082c5d6a352186e1be9464c5f31db81378824",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      188,
			createdBy: `/bin/sh -c mkdir -p /run/systemd && echo 'docker' > /run/systemd/container`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:e2f93437f92e69c54acb27971690e8fe49ba75783cc2e6d5b0efbaa971d73fba",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      2922771,
			createdBy: `RUN |1 TARGETPLATFORM=linux/amd64 /bin/sh -c apt-get update && apt-get upgrade -y     && apt-get install --no-install-recommends -y ca-certificates     && apt-get autoremove -y && apt-get clean     && rm -rf /tmp/* /var/tmp/*     && rm -rf /var/lib/apt/lists/* # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:21cb341b2283d5501142f9e4f9d1b1941138ccc0777b8914b18f842b42d1571c",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      120,
			createdBy: `RUN |1 TARGETPLATFORM=linux/amd64 /bin/sh -c mkdir -p /etc/envoy # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      21729278,
			createdBy: `ADD linux/amd64/build_release_stripped/* /usr/local/bin/ # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:3e05f50f195e6d16485c6a693092169b274d399d3cce98a87dd36c007a6911c3",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      749,
			createdBy: `ADD configs/envoyproxy_io_proxy.yaml /etc/envoy/envoy.yaml # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      3500,
			createdBy: `ADD linux/amd64/build_release/su-exec /usr/local/bin/ # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:2fb3fe4b571942f3d49d9c0ab84550cfa3843936278ce4e58dba28934efeff72",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      1467,
			createdBy: `RUN |2 TARGETPLATFORM=linux/amd64 ENVOY_BINARY_SUFFIX=_stripped /bin/sh -c chown root:root /usr/local/bin/su-exec && adduser --group --system envoy # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      490,
			createdBy: `COPY ci/docker-entrypoint.sh / # buildkit`,
		},
	},
}

func TestNewImage_LinuxAmd64(t *testing.T) {
	var i imageManifestV1
	require.NoError(t, json.Unmarshal(linuxAmd64VndDockerImageManifestV1Json, &i))
	var c imageConfigV1
	require.NoError(t, json.Unmarshal(linuxAmd64VndDockerImageConfigV1Json, &c))
	i.URL = "https://test/v2/user/repo/manifests/sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f"

	require.Equal(t, imageLinuxAmd64, newImage("https://test/v2/user/repo", &i, &c))
}

var imageLinuxArm64 = image{
	url:      "https://test/v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c",
	platform: "linux/arm64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "https://test/v2/user/repo/blobs/sha256:673aeee5c81c892477834e2b5e55575f16bfd52d9b841a1d8c524fb3805ee960",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      23703698,
			createdBy: `/bin/sh -c #(nop) ADD file:5f7cb4b44f843eaef6ae7ddb75dfc228a33d20cd974074ca23c1bb2cad7f77ad in / `,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:018b2790219d2003c0d437e634927887ee5cc3d8f985d7459adc5b2ff62d003f",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      851,
			createdBy: `/bin/sh -c set -xe 		&& echo '#!/bin/sh' > /usr/sbin/policy-rc.d 	&& echo 'exit 101' >> /usr/sbin/policy-rc.d 	&& chmod +x /usr/sbin/policy-rc.d 		&& dpkg-divert --local --rename --add /sbin/initctl 	&& cp -a /usr/sbin/policy-rc.d /sbin/initctl 	&& sed -i 's/^exit.*/exit 0/' /sbin/initctl 		&& echo 'force-unsafe-io' > /etc/dpkg/dpkg.cfg.d/docker-apt-speedup 		&& echo 'DPkg::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' > /etc/apt/apt.conf.d/docker-clean 	&& echo 'APT::Update::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' >> /etc/apt/apt.conf.d/docker-clean 	&& echo 'Dir::Cache::pkgcache ""; Dir::Cache::srcpkgcache "";' >> /etc/apt/apt.conf.d/docker-clean 		&& echo 'Acquire::Languages "none";' > /etc/apt/apt.conf.d/docker-no-languages 		&& echo 'Acquire::GzipIndexes "true"; Acquire::CompressionTypes::Order:: "gz";' > /etc/apt/apt.conf.d/docker-gzip-indexes 		&& echo 'Apt::AutoRemove::SuggestsImportant "false";' > /etc/apt/apt.conf.d/docker-autoremove-suggests`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:509c77ce92ade89fbf09fe03b167023be51bf5a0c14c00487fa7a9ee33b55fc3",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      187,
			createdBy: `/bin/sh -c mkdir -p /run/systemd && echo 'docker' > /run/systemd/container`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:1cfa500dd01835df61b905a437de186592fa2adf6d6a3694a26c13f76c72b1f6",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      2617240,
			createdBy: `RUN |1 TARGETPLATFORM=linux/arm64 /bin/sh -c apt-get update && apt-get upgrade -y     && apt-get install --no-install-recommends -y ca-certificates     && apt-get autoremove -y && apt-get clean     && rm -rf /tmp/* /var/tmp/*     && rm -rf /var/lib/apt/lists/* # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:57227c32adb08b6f11b734f43a3c621a25a35833d2eaff6047612deffabea67f",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      120,
			createdBy: `RUN |1 TARGETPLATFORM=linux/arm64 /bin/sh -c mkdir -p /etc/envoy # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:97c59091ec632eb43a1f8ae51f48200b97a580b9fbf0c591ad5cccd12d2bd573",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      19994790,
			createdBy: `ADD linux/arm64/build_release_stripped/* /usr/local/bin/ # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:2a7ca8a5ead0b680d1e00675e8f0a3ee864e64173e7150fd056bd72659f69bd6",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      746,
			createdBy: `ADD configs/envoyproxy_io_proxy.yaml /etc/envoy/envoy.yaml # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:af66acd072fe6384d76fe0f86ccf256a9a6ae9c6cb8b2b38c9ea4241cb92aeca",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      3888,
			createdBy: `ADD linux/arm64/build_release/su-exec /usr/local/bin/ # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:f21ff7be3ac20eb86e923b81c6735b98f980e793bb88db26716944bb5f8730f0",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      1460,
			createdBy: `RUN |2 TARGETPLATFORM=linux/arm64 ENVOY_BINARY_SUFFIX=_stripped /bin/sh -c chown root:root /usr/local/bin/su-exec && adduser --group --system envoy # buildkit`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      490,
			createdBy: `COPY ci/docker-entrypoint.sh / # buildkit`,
		},
	},
}

func TestNewImage_LinuxArm64(t *testing.T) {
	var image imageManifestV1
	require.NoError(t, json.Unmarshal(linuxArm64VndDockerImageManifestV1Json, &image))
	var config imageConfigV1
	require.NoError(t, json.Unmarshal(linuxArm64VndDockerImageConfigV1Json, &config))
	image.URL = "https://test/v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c"

	for i := range imageLinuxArm64.filesystemLayers {
		require.Equal(t, imageLinuxArm64.filesystemLayers[i], newImage("https://test/v2/user/repo", &image, &config).FilesystemLayer(i))
	}
}

//go:embed testdata/json/wasm-compat-vnd.oci.image.config.v1.json
var wasmCompatVndOciImageConfigV1Json []byte

func TestImageConfigV1_WasmCompat(t *testing.T) {
	var v imageConfigV1
	require.NoError(t, json.Unmarshal(wasmCompatVndOciImageConfigV1Json, &v))

	require.Equal(t, imageConfigV1{
		Architecture: "amd64",
		OS:           "linux",
		History:      []historyV1{{CreatedBy: "COPY plugin.wasm ./ # buildkit"}},
	}, v)
}

//go:embed testdata/json/wasm-compat-vnd.oci.image.manifest.v1.json
var wasmCompatVndOciImageManifestV1Json []byte

func TestImageManifestV1_WasmCompat(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(wasmCompatVndOciImageManifestV1Json, &v))

	require.Equal(t, imageManifestV1{
		Config: descriptorV1{
			MediaType: api.MediaTypeDockerContainerImage,
			Digest:    "sha256:453ac05d32d4a692870ff11cbee61edb7f05c4223ab772d10aaa37d5c150037a",
		},
		Layers: []descriptorV1{
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:d5e23ba78042fb166c603420339d92abb56a79bc8b689f4c84c96232a66be157",
				Size:      116164,
			},
		},
	}, v)
}

var imageWasmCompat = image{
	url:      "https://test/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672",
	platform: "linux/amd64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "https://test/v2/user/repo/blobs/sha256:d5e23ba78042fb166c603420339d92abb56a79bc8b689f4c84c96232a66be157",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      116164,
			createdBy: "COPY plugin.wasm ./ # buildkit",
		},
	},
}

func TestNewImage_WasmCompat(t *testing.T) {
	var i imageManifestV1
	require.NoError(t, json.Unmarshal(wasmCompatVndOciImageManifestV1Json, &i))
	var c imageConfigV1
	require.NoError(t, json.Unmarshal(wasmCompatVndOciImageConfigV1Json, &c))
	i.URL = "https://test/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672"
	require.Equal(t, imageWasmCompat, newImage("https://test/v2/user/repo", &i, &c))
}

//go:embed testdata/json/windows-vnd.docker.container.image.v1.json
var windowsVndDockerImageConfigV1Json []byte

func TestImageConfigV1_Windows(t *testing.T) {
	var v imageConfigV1
	require.NoError(t, json.Unmarshal(windowsVndDockerImageConfigV1Json, &v))

	require.Equal(t, imageConfigV1{
		Architecture: "amd64",
		OS:           "windows",
		OSVersion:    "10.0.17763.1879",
		History: []historyV1{
			{`Apply image 1809-RTM-amd64`, false},
			{`Install update ltsc2019-amd64`, false},
			{`cmd /S /C mkdir "C:\\Program\ Files\\envoy"`, false},
			{`cmd /S /C setx path "%path%;c:\Program Files\envoy"`, false},
			{`cmd /S /C #(nop) ADD file:61df7bfb8255c0673d4ed25f961df5121141ee800202081e549fc36828624577 in C:\Program Files\envoy\ `, false},
			{`cmd /S /C mkdir "C:\\ProgramData\\envoy"`, false},
			{`cmd /S /C #(nop) ADD file:59ef68147ad4a3f10999e2e334cf60397fbcc6501b3949dd811afd7b8f03ca43 in C:\ProgramData\envoy\envoy.yaml `, false},
			{`cmd /S /C powershell -Command "(cat C:\ProgramData\envoy\envoy.yaml -raw) -replace '/tmp/','C:\Windows\Temp\' | Set-Content -Encoding Ascii C:\ProgramData\envoy\envoy.yaml"`, false},
			{`cmd /S /C #(nop)  EXPOSE 10000`, false},
			{`cmd /S /C #(nop) COPY file:4e78f00367722220f515590585490fc6d785cc05e3a59a54f965431fa3ef374e in C:\ `, false},
			{`cmd /S /C #(nop)  ENTRYPOINT ["C:/docker-entrypoint.bat"]`, false},
			{`cmd /S /C #(nop)  CMD ["envoy.exe" "-c" "C:\\ProgramData\\envoy\\envoy.yaml"]`, false},
		},
	}, v)
}

//go:embed testdata/json/windows-vnd.docker.distribution.manifest.v2.json
var windowsVndDockerImageManifestV1Json []byte

func TestImageManifestV1_Windows(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(windowsVndDockerImageManifestV1Json, &v))

	require.Equal(t, imageManifestV1{
		Config: descriptorV1{
			MediaType: api.MediaTypeDockerContainerImage,
			Digest:    "sha256:00378fa4979bfcc7d1f5d33bb8cebe526395021801f9e233f8909ffc25a6f630",
		},
		Layers: []descriptorV1{
			{
				MediaType: "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
				Digest:    "sha256:4612f6d0b889cad0ed0292fae3a0b0c8a9e49aff6dea8eb049b2386d9b07986f",
				Size:      1718332879,
			},
			{
				MediaType: "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip",
				Digest:    "sha256:399f118dfaa9a753e98d128238b944432c7bcabea88a2998a6efbbece28ed303",
				Size:      751421005,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:47916aee02007e0e175e80deb2938cf8f95457b9abb555bd44dc461680dc552c",
				Size:      323887,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:ba79ee4428b5ceec3026664126a146fd8c1041b478f3018ec0c90b78d7fe6355",
				Size:      331919,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:fd103a6c37aad8ffeaef6521612ed5a5153b104fffdb8bf3b6cf3d0beaaa49c4",
				Size:      12217107,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:0fcfdc906e922391139a1c2d8f5d600066fa3b21c720a4024831471e2a8f0011",
				Size:      337530,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:f5ece8fbad694f5d1169c17ddd4217265cdf3dd886b71a8e9144f8b00e22de07",
				Size:      2410,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:8d3db7768af4371ec3f749f6816c8450687e276a883b8ca626a1fc1402fd32e0",
				Size:      419457,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:f0b13e108f65feef6ee7b28a639a516aa37082bca3e0ac332bcde1e97e095b6b",
				Size:      1303,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:9e17bb8cfb82c53b1793341a2dfb555e63088b1594d81d2b01106fae9a8aa60b",
				Size:      1745,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:30188a58a9ae8bd6cbfc36a6ba873a1a8cfe5a50993fc982844b935fa2724126",
				Size:      1327,
			},
			{
				MediaType: api.MediaTypeDockerImageLayer,
				Digest:    "sha256:ce93263143f489be1ca45bbda23e98dc97445fc9b3d53a7ffd4f7a7eb25889fc",
				Size:      1334,
			},
		},
	}, v)
}

var imageWindows = image{
	url:      "https://test/v2/user/repo/manifests/v1.0",
	platform: "windows/amd64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "https://test/v2/user/repo/blobs/sha256:47916aee02007e0e175e80deb2938cf8f95457b9abb555bd44dc461680dc552c",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      323887,
			createdBy: `cmd /S /C mkdir "C:\\Program\ Files\\envoy"`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:ba79ee4428b5ceec3026664126a146fd8c1041b478f3018ec0c90b78d7fe6355",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      331919,
			createdBy: `cmd /S /C setx path "%path%;c:\Program Files\envoy"`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:fd103a6c37aad8ffeaef6521612ed5a5153b104fffdb8bf3b6cf3d0beaaa49c4",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      12217107,
			createdBy: `cmd /S /C #(nop) ADD file:61df7bfb8255c0673d4ed25f961df5121141ee800202081e549fc36828624577 in C:\Program Files\envoy\ `,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:0fcfdc906e922391139a1c2d8f5d600066fa3b21c720a4024831471e2a8f0011",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      337530,
			createdBy: `cmd /S /C mkdir "C:\\ProgramData\\envoy"`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:f5ece8fbad694f5d1169c17ddd4217265cdf3dd886b71a8e9144f8b00e22de07",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      2410,
			createdBy: `cmd /S /C #(nop) ADD file:59ef68147ad4a3f10999e2e334cf60397fbcc6501b3949dd811afd7b8f03ca43 in C:\ProgramData\envoy\envoy.yaml `,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:8d3db7768af4371ec3f749f6816c8450687e276a883b8ca626a1fc1402fd32e0",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      419457,
			createdBy: `cmd /S /C powershell -Command "(cat C:\ProgramData\envoy\envoy.yaml -raw) -replace '/tmp/','C:\Windows\Temp\' | Set-Content -Encoding Ascii C:\ProgramData\envoy\envoy.yaml"`,
		},
		{
			url:       "https://test/v2/user/repo/blobs/sha256:9e17bb8cfb82c53b1793341a2dfb555e63088b1594d81d2b01106fae9a8aa60b",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      1745,
			createdBy: `cmd /S /C #(nop) COPY file:4e78f00367722220f515590585490fc6d785cc05e3a59a54f965431fa3ef374e in C:\ `,
		},
	},
}

func TestNewImage_Windows(t *testing.T) {
	var i imageManifestV1
	require.NoError(t, json.Unmarshal(windowsVndDockerImageManifestV1Json, &i))
	var c imageConfigV1
	require.NoError(t, json.Unmarshal(windowsVndDockerImageConfigV1Json, &c))
	i.URL = "https://test/v2/user/repo/manifests/v1.0"
	require.Equal(t, imageWindows, newImage("https://test/v2/user/repo", &i, &c))
}

//go:embed testdata/json/trivy-vnd.oci.image.manifest.v1.json
var trivyVndOciImageManifestV1Json []byte

var trivyVndOciUnknownConfigV1Json = []byte("{}")

func TestImageManifestV1_Trivy(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(trivyVndOciImageManifestV1Json, &v))

	require.Equal(t, imageManifestV1{
		Config: descriptorV1{
			MediaType: api.MediaTypeUnknownImageConfig,
			Digest:    "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			Size:      2,
		},
		Layers: []descriptorV1{
			{
				MediaType: api.MediaTypeModuleWasmImageLayer,
				Digest:    "sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b",
				Size:      460018,
				Annotations: map[string]string{
					opencontainersImageTitle: "wordpress.wasm",
				},
			},
		},
	}, v)
}

var imageTrivy = image{
	url:      "https://test/v2/user/repo/manifests/v1.0",
	platform: "", // unknown
	filesystemLayers: []filesystemLayer{
		{
			url:       "https://test/v2/user/repo/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b",
			mediaType: api.MediaTypeModuleWasmImageLayer,
			size:      460018,
			fileName:  "wordpress.wasm",
		},
	},
}

func TestNewImage_Trivy(t *testing.T) {
	var i imageManifestV1
	require.NoError(t, json.Unmarshal(trivyVndOciImageManifestV1Json, &i))
	var c imageConfigV1
	require.NoError(t, json.Unmarshal(trivyVndOciUnknownConfigV1Json, &c))
	i.URL = imageTrivy.url
	require.Equal(t, imageTrivy, newImage("https://test/v2/user/repo", &i, &c))
}

//go:embed testdata/json/krustlet-vnd.oci.image.manifest.v1.json
var krustletVndOciImageManifestV1Json []byte

var krustletVndWasmConfigV1Json = []byte("{}")

func TestImageManifestV1_Krustlet(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(krustletVndOciImageManifestV1Json, &v))

	require.Equal(t, imageManifestV1{
		Config: descriptorV1{
			MediaType: api.MediaTypeWasmImageConfig,
			Digest:    "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
			Size:      2,
		},
		Layers: []descriptorV1{
			{
				MediaType: api.MediaTypeWasmImageLayer,
				Digest:    "sha256:f9c91f4c280ab92aff9eb03b279c4774a80b84428741ab20855d32004b2b983f",
				Size:      1615998,
				Annotations: map[string]string{
					opencontainersImageTitle: "module.wasm",
				},
			},
		},
	}, v)
}

// actual example is from ghcr.io/krustlet/oci-distribution/hello-wasm:v1
var imageKrustlet = image{
	url:      "https://test/v2/user/repo/manifests/v1.0",
	platform: "", // unknown
	filesystemLayers: []filesystemLayer{
		{
			url:       "https://test/v2/user/repo/blobs/sha256:f9c91f4c280ab92aff9eb03b279c4774a80b84428741ab20855d32004b2b983f",
			mediaType: api.MediaTypeWasmImageLayer,
			size:      1615998,
			fileName:  "module.wasm",
		},
	},
}

func TestNewImage_Krustlet(t *testing.T) {
	var i imageManifestV1
	require.NoError(t, json.Unmarshal(krustletVndOciImageManifestV1Json, &i))
	var c imageConfigV1
	require.NoError(t, json.Unmarshal(krustletVndWasmConfigV1Json, &c))
	i.URL = imageKrustlet.url
	require.Equal(t, imageKrustlet, newImage("https://test/v2/user/repo", &i, &c))
}

// TestSkipCreatedByPattern ensures fallback logic works when historyV1.EmptyLayer is not set.
func TestSkipCreatedByPattern(t *testing.T) {
	tests := []struct {
		name, createdBy      string
		emptyLayer, expected bool
	}{
		{
			name:     "doesn't skip empty createdBy",
			expected: false,
		},
		{
			name:      "doesn't skip ADD",
			createdBy: `ADD linux/amd64/build_release/su-exec /usr/local/bin/ # buildkit`,
			expected:  false,
		},
		{
			name:      "doesn't skip ADD (windows)",
			createdBy: `cmd /S /C #(nop) ADD file:61df7bfb8255c0673d4ed25f961df5121141ee800202081e549fc36828624577 in C:\Program Files\envoy\ `,
			expected:  false,
		},
		{
			name:      "doesn't skip COPY",
			createdBy: `COPY ci/docker-entrypoint.sh / # buildkit`,
			expected:  false,
		},
		{
			name:      "doesn't skip COPY (windows)",
			createdBy: `cmd /S /C #(nop) COPY file:4e78f00367722220f515590585490fc6d785cc05e3a59a54f965431fa3ef374e in C:\ `,
			expected:  false,
		},
		{
			name:      "doesn't skip RUN",
			createdBy: `/bin/sh -c mkdir -p /run/systemd && echo 'docker' > /run/systemd/container`,
			expected:  false,
		},
		{
			name:      "doesn't skip RUN (windows)",
			createdBy: `cmd /S /C mkdir "C:\\ProgramData\\envoy"`,
			expected:  false,
		},
		{
			name:      "skips ignored Docker directive (windows)", // windows doesn't always use emptyLayer
			createdBy: `cmd /S /C #(nop)  EXPOSE 10000`,           // extra spaces
			expected:  true,
		},
	}

	for _, tc := range tests {
		tc := tc // pin! see https://github.com/kyoh86/scopelint for why

		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, skipCreatedByPattern.MatchString(tc.createdBy))
		})
	}
}
