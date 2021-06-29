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

	"github.com/tetratelabs/car/internal"
)

//go:embed testdata/homebrew-vnd.oci.image.config.v1.json
var homebrewVndOciImageConfigV1Json []byte

func TestImageConfigV1_Homebrew(t *testing.T) {
	var v imageConfigV1
	require.NoError(t, json.Unmarshal(homebrewVndOciImageConfigV1Json, &v))

	require.Equal(t, imageConfigV1{
		Architecture: internal.ArchAmd64,
		OS:           internal.OSDarwin,
		OSVersion:    "macOS 11.3",
	}, v)
}

//go:embed testdata/homebrew-vnd.oci.image.index.v1.json
var homebrewVndOciImageIndexV1Json []byte

func TestImageIndexV1_Homebrew(t *testing.T) {
	var v imageIndexV1
	require.NoError(t, json.Unmarshal(homebrewVndOciImageIndexV1Json, &v))

	require.Equal(t, imageIndexV1{
		Manifests: []imageManifestReferenceV1{
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:0da7ea4ca0f3615ace3b2223248e0baed539223df62d33d4c1a1e23346329057",
				Platform:  platformV1{internal.ArchAmd64, internal.OSDarwin, "macOS 10.15.7"},
			},
			{
				MediaType: "application/vnd.oci.image.manifest.v1+json",
				Digest:    "sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672",
				Platform:  platformV1{internal.ArchAmd64, internal.OSDarwin, "macOS 11.3"},
			},
		},
	}, v)
}

//go:embed testdata/homebrew-vnd.oci.image.manifest.v1.json
var homebrewVndOciImageManifestV1Json []byte

func TestImageManifestV1_Homebrew(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(homebrewVndOciImageManifestV1Json, &v))

	require.Equal(t, imageManifestV1{
		Config: descriptorV1{
			MediaType: "application/vnd.oci.image.config.v1+json",
			Digest:    "sha256:a7f8bac78026ae40545531454c2ef4df75ec3de1c60f1d6923142fe4e44daf8a",
			Size:      222,
		},
		Layers: []descriptorV1{
			{mediaTypeOCILayer, "sha256:d03fb86b48336c8d3c0f3711cfc3df3557f9fb33c966ceb1caecae1653935e90", 29405739},
		},
	}, v)
}

//go:embed testdata/linux-vnd.docker.container.image.v1.json
var linuxVndDockerImageConfigV1Json []byte

func TestImageConfigV1_Linux(t *testing.T) {
	var v imageConfigV1
	require.NoError(t, json.Unmarshal(linuxVndDockerImageConfigV1Json, &v))

	require.Equal(t, imageConfigV1{
		Architecture: internal.ArchAmd64,
		OS:           internal.OSLinux,
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

//go:embed testdata/linux-vnd.docker.distribution.manifest.list.v2.json
var linuxVndDockerImageIndexV1Json []byte

func TestImageIndexV1_Linux(t *testing.T) {
	var v imageIndexV1
	require.NoError(t, json.Unmarshal(linuxVndDockerImageIndexV1Json, &v))

	require.Equal(t, imageIndexV1{
		Manifests: []imageManifestReferenceV1{
			{
				MediaType: "application/vnd.docker.distribution.manifest.v2+json",
				Digest:    "sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c",
				Platform:  platformV1{internal.ArchArm64, internal.OSLinux, ""},
			},
			{
				MediaType: "application/vnd.docker.distribution.manifest.v2+json",
				Digest:    "sha256:4e07f3bd88fb4a468d5551c21eb05f625b0efe9ee00ae25d3ffb87c0f563693f",
				Platform:  platformV1{internal.ArchAmd64, internal.OSLinux, ""},
			},
		},
	}, v)
}

//go:embed testdata/linux-vnd.docker.distribution.manifest.v2.json
var linuxVndDockerImageManifestV1Json []byte

func TestImageManifestV1_Linux(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(linuxVndDockerImageManifestV1Json, &v))

	require.Equal(t, imageManifestV1{
		Config: descriptorV1{
			MediaType: "application/vnd.docker.container.image.v1+json",
			Digest:    "sha256:33655f17f09318801873b70f89c1596ce38f41f6c074e2343d26e9b425f939ec",
			Size:      5298,
		},
		Layers: []descriptorV1{
			{mediaTypeDockerLayer, "sha256:01bf7da0a88c9e37ae418d17c0aeed0621524848d80ccb9e38c67e7ab8e11928", 26697009},
			{mediaTypeDockerLayer, "sha256:f3b4a5f15c7a0722b4f22e61b5387317eaf2602c27ffb2bceac9a25f19fbd156", 852},
			{mediaTypeDockerLayer, "sha256:57ffbe87baa135002dddb7a7460082c5d6a352186e1be9464c5f31db81378824", 188},
			{mediaTypeDockerLayer, "sha256:e2f93437f92e69c54acb27971690e8fe49ba75783cc2e6d5b0efbaa971d73fba", 2922771},
			{mediaTypeDockerLayer, "sha256:21cb341b2283d5501142f9e4f9d1b1941138ccc0777b8914b18f842b42d1571c", 120},
			{mediaTypeDockerLayer, "sha256:15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2", 21729278},
			{mediaTypeDockerLayer, "sha256:3e05f50f195e6d16485c6a693092169b274d399d3cce98a87dd36c007a6911c3", 749},
			{mediaTypeDockerLayer, "sha256:1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81", 3500},
			{mediaTypeDockerLayer, "sha256:2fb3fe4b571942f3d49d9c0ab84550cfa3843936278ce4e58dba28934efeff72", 1467},
			{mediaTypeDockerLayer, "sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb", 490},
		},
	}, v)
}

func TestNewImage_Linux(t *testing.T) {
	var i imageManifestV1
	require.NoError(t, json.Unmarshal(linuxVndDockerImageManifestV1Json, &i))
	var c imageConfigV1
	require.NoError(t, json.Unmarshal(linuxVndDockerImageConfigV1Json, &c))

	require.Equal(t, &internal.Image{
		Platform: internal.OSLinux + "/" + internal.ArchAmd64,
		FilesystemLayers: []*internal.FilesystemLayer{
			{
				URL:       "/blobs/sha256:01bf7da0a88c9e37ae418d17c0aeed0621524848d80ccb9e38c67e7ab8e11928",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      26697009,
				CreatedBy: `/bin/sh -c #(nop) ADD file:d7fa3c26651f9204a5629287a1a9a6e7dc6a0bc6eb499e82c433c0c8f67ff46b in / `,
			},
			{
				URL:       "/blobs/sha256:f3b4a5f15c7a0722b4f22e61b5387317eaf2602c27ffb2bceac9a25f19fbd156",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      852,
				CreatedBy: `/bin/sh -c set -xe 		&& echo '#!/bin/sh' > /usr/sbin/policy-rc.d 	&& echo 'exit 101' >> /usr/sbin/policy-rc.d 	&& chmod +x /usr/sbin/policy-rc.d 		&& dpkg-divert --local --rename --add /sbin/initctl 	&& cp -a /usr/sbin/policy-rc.d /sbin/initctl 	&& sed -i 's/^exit.*/exit 0/' /sbin/initctl 		&& echo 'force-unsafe-io' > /etc/dpkg/dpkg.cfg.d/docker-apt-speedup 		&& echo 'DPkg::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' > /etc/apt/apt.conf.d/docker-clean 	&& echo 'APT::Update::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' >> /etc/apt/apt.conf.d/docker-clean 	&& echo 'Dir::Cache::pkgcache ""; Dir::Cache::srcpkgcache "";' >> /etc/apt/apt.conf.d/docker-clean 		&& echo 'Acquire::Languages "none";' > /etc/apt/apt.conf.d/docker-no-languages 		&& echo 'Acquire::GzipIndexes "true"; Acquire::CompressionTypes::Order:: "gz";' > /etc/apt/apt.conf.d/docker-gzip-indexes 		&& echo 'Apt::AutoRemove::SuggestsImportant "false";' > /etc/apt/apt.conf.d/docker-autoremove-suggests`,
			},
			{
				URL:       "/blobs/sha256:57ffbe87baa135002dddb7a7460082c5d6a352186e1be9464c5f31db81378824",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      188,
				CreatedBy: `/bin/sh -c mkdir -p /run/systemd && echo 'docker' > /run/systemd/container`,
			}, {
				URL:       "/blobs/sha256:e2f93437f92e69c54acb27971690e8fe49ba75783cc2e6d5b0efbaa971d73fba",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      2922771,
				CreatedBy: `RUN |1 TARGETPLATFORM=linux/amd64 /bin/sh -c apt-get update && apt-get upgrade -y     && apt-get install --no-install-recommends -y ca-certificates     && apt-get autoremove -y && apt-get clean     && rm -rf /tmp/* /var/tmp/*     && rm -rf /var/lib/apt/lists/* # buildkit`,
			}, {
				URL:       "/blobs/sha256:21cb341b2283d5501142f9e4f9d1b1941138ccc0777b8914b18f842b42d1571c",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      120,
				CreatedBy: `RUN |1 TARGETPLATFORM=linux/amd64 /bin/sh -c mkdir -p /etc/envoy # buildkit`,
			}, {
				URL:       "/blobs/sha256:15a7c58f96c57b941a56cbf1bdd525cdef1773a7671c52b7039047a1941105c2",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      21729278,
				CreatedBy: `ADD linux/amd64/build_release_stripped/* /usr/local/bin/ # buildkit`,
			}, {
				URL:       "/blobs/sha256:3e05f50f195e6d16485c6a693092169b274d399d3cce98a87dd36c007a6911c3",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      749,
				CreatedBy: `ADD configs/envoyproxy_io_proxy.yaml /etc/envoy/envoy.yaml # buildkit`,
			}, {
				URL:       "/blobs/sha256:1b68df344f018b7cdd39908b93b6d60792a414cbf47975f7606a18bd603e6a81",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      3500,
				CreatedBy: `ADD linux/amd64/build_release/su-exec /usr/local/bin/ # buildkit`,
			}, {
				URL:       "/blobs/sha256:2fb3fe4b571942f3d49d9c0ab84550cfa3843936278ce4e58dba28934efeff72",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      1467,
				CreatedBy: `RUN |2 TARGETPLATFORM=linux/amd64 ENVOY_BINARY_SUFFIX=_stripped /bin/sh -c chown root:root /usr/local/bin/su-exec && adduser --group --system envoy # buildkit`,
			}, {
				URL:       "/blobs/sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      490,
				CreatedBy: `COPY ci/docker-entrypoint.sh / # buildkit`,
			}},
	}, newImage("", &i, &c))
}

//go:embed testdata/windows-vnd.docker.container.image.v1.json
var windowsVndDockerImageConfigV1Json []byte

func TestImageConfigV1_Windows(t *testing.T) {
	var v imageConfigV1
	require.NoError(t, json.Unmarshal(windowsVndDockerImageConfigV1Json, &v))

	require.Equal(t, imageConfigV1{
		Architecture: internal.ArchAmd64,
		OS:           internal.OSWindows,
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

//go:embed testdata/windows-vnd.docker.distribution.manifest.v2.json
var windowsVndDockerImageManifestV1Json []byte

func TestImageManifestV1_Windows(t *testing.T) {
	var v imageManifestV1
	require.NoError(t, json.Unmarshal(windowsVndDockerImageManifestV1Json, &v))

	require.Equal(t, imageManifestV1{
		Config: descriptorV1{
			MediaType: "application/vnd.docker.container.image.v1+json",
			Digest:    "sha256:00378fa4979bfcc7d1f5d33bb8cebe526395021801f9e233f8909ffc25a6f630",
			Size:      3744,
		},
		Layers: []descriptorV1{
			{"application/vnd.docker.image.rootfs.foreign.diff.tar.gzip", "sha256:4612f6d0b889cad0ed0292fae3a0b0c8a9e49aff6dea8eb049b2386d9b07986f", 1718332879},
			{"application/vnd.docker.image.rootfs.foreign.diff.tar.gzip", "sha256:399f118dfaa9a753e98d128238b944432c7bcabea88a2998a6efbbece28ed303", 751421005},
			{mediaTypeDockerLayer, "sha256:47916aee02007e0e175e80deb2938cf8f95457b9abb555bd44dc461680dc552c", 323887},
			{mediaTypeDockerLayer, "sha256:ba79ee4428b5ceec3026664126a146fd8c1041b478f3018ec0c90b78d7fe6355", 331919},
			{mediaTypeDockerLayer, "sha256:fd103a6c37aad8ffeaef6521612ed5a5153b104fffdb8bf3b6cf3d0beaaa49c4", 12217107},
			{mediaTypeDockerLayer, "sha256:0fcfdc906e922391139a1c2d8f5d600066fa3b21c720a4024831471e2a8f0011", 337530},
			{mediaTypeDockerLayer, "sha256:f5ece8fbad694f5d1169c17ddd4217265cdf3dd886b71a8e9144f8b00e22de07", 2410},
			{mediaTypeDockerLayer, "sha256:8d3db7768af4371ec3f749f6816c8450687e276a883b8ca626a1fc1402fd32e0", 419457},
			{mediaTypeDockerLayer, "sha256:f0b13e108f65feef6ee7b28a639a516aa37082bca3e0ac332bcde1e97e095b6b", 1303},
			{mediaTypeDockerLayer, "sha256:9e17bb8cfb82c53b1793341a2dfb555e63088b1594d81d2b01106fae9a8aa60b", 1745},
			{mediaTypeDockerLayer, "sha256:30188a58a9ae8bd6cbfc36a6ba873a1a8cfe5a50993fc982844b935fa2724126", 1327},
			{mediaTypeDockerLayer, "sha256:ce93263143f489be1ca45bbda23e98dc97445fc9b3d53a7ffd4f7a7eb25889fc", 1334},
		},
	}, v)
}

func TestNewImage_Windows(t *testing.T) {
	var i imageManifestV1
	require.NoError(t, json.Unmarshal(windowsVndDockerImageManifestV1Json, &i))
	var c imageConfigV1
	require.NoError(t, json.Unmarshal(windowsVndDockerImageConfigV1Json, &c))

	require.Equal(t, &internal.Image{
		Platform: internal.OSWindows + "/" + internal.ArchAmd64,
		FilesystemLayers: []*internal.FilesystemLayer{
			{
				URL:       "/blobs/sha256:47916aee02007e0e175e80deb2938cf8f95457b9abb555bd44dc461680dc552c",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      323887,
				CreatedBy: `cmd /S /C mkdir "C:\\Program\ Files\\envoy"`,
			},
			{
				URL:       "/blobs/sha256:ba79ee4428b5ceec3026664126a146fd8c1041b478f3018ec0c90b78d7fe6355",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      331919,
				CreatedBy: `cmd /S /C setx path "%path%;c:\Program Files\envoy"`,
			},
			{
				URL:       "/blobs/sha256:fd103a6c37aad8ffeaef6521612ed5a5153b104fffdb8bf3b6cf3d0beaaa49c4",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      12217107,
				CreatedBy: `cmd /S /C #(nop) ADD file:61df7bfb8255c0673d4ed25f961df5121141ee800202081e549fc36828624577 in C:\Program Files\envoy\ `,
			},
			{
				URL:       "/blobs/sha256:0fcfdc906e922391139a1c2d8f5d600066fa3b21c720a4024831471e2a8f0011",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      337530,
				CreatedBy: `cmd /S /C mkdir "C:\\ProgramData\\envoy"`,
			},
			{
				URL:       "/blobs/sha256:f5ece8fbad694f5d1169c17ddd4217265cdf3dd886b71a8e9144f8b00e22de07",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      2410,
				CreatedBy: `cmd /S /C #(nop) ADD file:59ef68147ad4a3f10999e2e334cf60397fbcc6501b3949dd811afd7b8f03ca43 in C:\ProgramData\envoy\envoy.yaml `,
			},
			{
				URL:       "/blobs/sha256:8d3db7768af4371ec3f749f6816c8450687e276a883b8ca626a1fc1402fd32e0",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      419457,
				CreatedBy: `cmd /S /C powershell -Command "(cat C:\ProgramData\envoy\envoy.yaml -raw) -replace '/tmp/','C:\Windows\Temp\' | Set-Content -Encoding Ascii C:\ProgramData\envoy\envoy.yaml"`,
			},
			{
				URL:       "/blobs/sha256:9e17bb8cfb82c53b1793341a2dfb555e63088b1594d81d2b01106fae9a8aa60b",
				MediaType: "application/vnd.docker.image.rootfs.diff.tar.gzip",
				Size:      1745,
				CreatedBy: `cmd /S /C #(nop) COPY file:4e78f00367722220f515590585490fc6d785cc05e3a59a54f965431fa3ef374e in C:\ `,
			},
		},
	}, newImage("", &i, &c))
}
