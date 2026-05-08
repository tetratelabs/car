// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

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

//go:embed testdata/json/homebrew-vnd.oci.image.index.v1.json
var homebrewVndOciImageIndexV1Json []byte

//go:embed testdata/json/homebrew-11.3-vnd.oci.image.manifest.v1.json
var homebrew113VndOciImageManifestV1Json []byte

//go:embed testdata/json/envoy-v1.38.0-linux-amd64-vnd.oci.image.config.v1.json
var linuxAmd64VndOciImageConfigV1Json []byte

//go:embed testdata/json/linux-arm64-vnd.docker.container.image.v1.json
var linuxArm64VndDockerImageConfigV1Json []byte

//go:embed testdata/json/linux-vnd.docker.distribution.manifest.list.v2.json
var linuxVndDockerImageIndexV1Json []byte

//go:embed testdata/json/envoy-v1.38.0-linux-amd64-vnd.oci.image.manifest.v1.json
var linuxAmd64VndOciImageManifestV1Json []byte

//go:embed testdata/json/linux-arm64-vnd.docker.distribution.manifest.v2.json
var linuxArm64VndDockerImageManifestV1Json []byte

//go:embed testdata/json/wasm-compat-vnd.oci.image.config.v1.json
var wasmCompatVndOciImageConfigV1Json []byte

//go:embed testdata/json/wasm-compat-vnd.oci.image.manifest.v1.json
var wasmCompatVndOciImageManifestV1Json []byte

//go:embed testdata/json/windows-vnd.docker.container.image.v1.json
var windowsVndDockerImageConfigV1Json []byte

//go:embed testdata/json/windows-vnd.docker.distribution.manifest.v2.json
var windowsVndDockerImageManifestV1Json []byte

//go:embed testdata/json/trivy-vnd.oci.image.manifest.v1.json
var trivyVndOciImageManifestV1Json []byte

//go:embed testdata/json/krustlet-vnd.oci.image.manifest.v1.json
var krustletVndOciImageManifestV1Json []byte

func TestImageConfigV1(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected imageConfigV1
	}{
		{
			name:  "homebrew",
			input: homebrew113VndOciImageConfigV1Json,
			expected: imageConfigV1{
				Architecture: "amd64",
				OS:           "darwin",
				OSVersion:    "macOS 11.3",
			},
		},
		{
			name:  "linux/amd64",
			input: linuxAmd64VndOciImageConfigV1Json,
			expected: imageConfigV1{
				Architecture: "amd64",
				OS:           "linux",
				History: []historyV1{
					{`/bin/sh -c #(nop)  ARG RELEASE`, true},
					{`/bin/sh -c #(nop)  ARG LAUNCHPAD_BUILD_ARCH`, true},
					{`/bin/sh -c #(nop)  LABEL org.opencontainers.image.version=22.04`, true},
					{`/bin/sh -c #(nop) ADD file:da2cd86408d9354e8bd817c8a4b8635a1d788cd20d0d70061ce02a173e8cf902 in / `, false},
					{`/bin/sh -c #(nop)  CMD ["/bin/bash"]`, true},
					{`ENV DEBIAN_FRONTEND=noninteractive`, true},
					{`EXPOSE [10000/tcp]`, true},
					{`CMD ["envoy" "-c" "/etc/envoy/envoy.yaml"]`, true},
					{`RUN /bin/sh -c mkdir -p /etc/envoy     && adduser --group --system envoy # buildkit`, false},
					{`ENTRYPOINT ["/docker-entrypoint.sh"]`, true},
					{`ADD VERSION.txt /etc/envoy # buildkit`, false},
					{`RUN /bin/sh -c apt-get -qq update     && apt-get -qq upgrade -y     && apt-get -qq install --no-install-recommends -y ca-certificates tzdata     && apt-get -qq autoremove -y # buildkit`, false},
					{`COPY --chown=0:0 --chmod=644 /etc/envoy/envoy.yaml /etc/envoy/envoy.yaml # buildkit`, false},
					{`COPY --chown=0:0 --chmod=755 /docker-entrypoint.sh / # buildkit`, false},
					{`COPY --chown=0:0 --chmod=755 /usr/local/bin/utils/su-exec /usr/local/bin/ # buildkit`, false},
					{`ARG ENVOY_BINARY=envoy`, true},
					{`ARG ENVOY_BINARY_PREFIX=`, true},
					{`COPY --chown=0:0 --chmod=755 /usr/local/bin/envoy /usr/local/bin/envoy # buildkit`, false},
					{`COPY --chown=0:0 --chmod=755 /usr/local/bin/envoy.* /usr/local/bin/ # buildkit`, false},
				},
			},
		},
		{
			name:  "wasm compat",
			input: wasmCompatVndOciImageConfigV1Json,
			expected: imageConfigV1{
				Architecture: "amd64",
				OS:           "linux",
				History:      []historyV1{{CreatedBy: "COPY plugin.wasm ./ # buildkit"}},
			},
		},
		{
			name:  "windows",
			input: windowsVndDockerImageConfigV1Json,
			expected: imageConfigV1{
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
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var v imageConfigV1
			require.NoError(t, json.Unmarshal(tc.input, &v))
			require.Equal(t, tc.expected, v)
		})
	}
}

func TestImageIndexV1(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected imageIndexV1
	}{
		{
			name:  "homebrew",
			input: homebrewVndOciImageIndexV1Json,
			expected: imageIndexV1{
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
			},
		},
		{
			name:  "linux",
			input: linuxVndDockerImageIndexV1Json,
			expected: imageIndexV1{
				Manifests: []*imageManifestReferenceV1{
					{
						MediaType: api.MediaTypeDockerManifest,
						Digest:    "sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c",
						Platform:  platformV1{"arm64", "linux", ""},
					},
					{
						MediaType: api.MediaTypeOCIImageManifest,
						Digest:    "sha256:a85fb4ac4bce750803b9db6306152db456864486e36f61559731f9033a9293f0",
						Platform:  platformV1{"amd64", "linux", ""},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var v imageIndexV1
			require.NoError(t, json.Unmarshal(tc.input, &v))
			require.Equal(t, tc.expected, v)
		})
	}
}

func TestImageManifestV1(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected imageManifestV1
	}{
		{
			name:  "homebrew",
			input: homebrew113VndOciImageManifestV1Json,
			expected: imageManifestV1{
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
			},
		},
		{
			name:  "linux/amd64",
			input: linuxAmd64VndOciImageManifestV1Json,
			expected: imageManifestV1{
				Config: descriptorV1{
					MediaType: api.MediaTypeOCIImageConfig,
					Digest:    "sha256:ba952938f316ef8eb82a35a75bc79232a57233d20ce5ee14d63a6640d4257508",
					Size:      4130,
				},
				Layers: []descriptorV1{
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:f63eb04151bcac21ad049f8d781b97b219aba392c5457907f8f3e88e43eb48ec",
						Size:      29736498,
					},
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:2fec5073439e939b3b5b81a12627aa526c622a6226bf2302bf611be47b69a39d",
						Size:      1402,
					},
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:d38c216951ea481b7c3a22ae3a167fe78d9181de6bfd5b34f0ffac94d5600d8f",
						Size:      174,
					},
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:716deb9edfe95f123ec27c622cc346a30ee609d2a15c6077034faf1216244263",
						Size:      2015696,
					},
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:1923f5bd1d1e8f51a40c592f81dcdb17bc40fa9e67581424c175daf410301f6f",
						Size:      790,
					},
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:75a2c6152282d303c7c5fd2ffa8cad9c1ac0e4a0a909a72d147b69feba64deff",
						Size:      484,
					},
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:955a5dc17c1971170dae1b619dd415d9d3d481fb82bc45f2a4d1da2d4288791c",
						Size:      4679,
					},
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:a94127188f8273b23ce07bdcb990b4ea3f0c145103578defc11ce1e477a744eb",
						Size:      36100758,
					},
					{
						MediaType: api.MediaTypeOCIImageLayer,
						Digest:    "sha256:4f4fb700ef54461cfa02571ae0db9a0dc1e0cdb5577484a6d75e68dc38e8acc1",
						Size:      32,
					},
				},
			},
		},
		{
			name:  "wasm compat",
			input: wasmCompatVndOciImageManifestV1Json,
			expected: imageManifestV1{
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
			},
		},
		{
			name:  "windows",
			input: windowsVndDockerImageManifestV1Json,
			expected: imageManifestV1{
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
			},
		},
		{
			name:  "trivy",
			input: trivyVndOciImageManifestV1Json,
			expected: imageManifestV1{
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
			},
		},
		{
			name:  "krustlet",
			input: krustletVndOciImageManifestV1Json,
			expected: imageManifestV1{
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
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var v imageManifestV1
			require.NoError(t, json.Unmarshal(tc.input, &v))
			require.Equal(t, tc.expected, v)
		})
	}
}

var imageHomebrew = image{
	url:      "http://test:5000/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672",
	platform: "darwin/amd64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:d03fb86b48336c8d3c0f3711cfc3df3557f9fb33c966ceb1caecae1653935e90",
			mediaType: "application/vnd.oci.image.layer.v1.tar+gzip",
			size:      29405739,
		},
	},
}

var imageLinuxAmd64 = image{
	url:      "http://test:5000/v2/user/repo/manifests/sha256:a85fb4ac4bce750803b9db6306152db456864486e36f61559731f9033a9293f0",
	platform: "linux/amd64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:f63eb04151bcac21ad049f8d781b97b219aba392c5457907f8f3e88e43eb48ec",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      29736498,
			createdBy: `/bin/sh -c #(nop) ADD file:da2cd86408d9354e8bd817c8a4b8635a1d788cd20d0d70061ce02a173e8cf902 in / `,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:2fec5073439e939b3b5b81a12627aa526c622a6226bf2302bf611be47b69a39d",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      1402,
			createdBy: `RUN /bin/sh -c mkdir -p /etc/envoy     && adduser --group --system envoy # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:d38c216951ea481b7c3a22ae3a167fe78d9181de6bfd5b34f0ffac94d5600d8f",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      174,
			createdBy: `ADD VERSION.txt /etc/envoy # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:716deb9edfe95f123ec27c622cc346a30ee609d2a15c6077034faf1216244263",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      2015696,
			createdBy: `RUN /bin/sh -c apt-get -qq update     && apt-get -qq upgrade -y     && apt-get -qq install --no-install-recommends -y ca-certificates tzdata     && apt-get -qq autoremove -y # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:1923f5bd1d1e8f51a40c592f81dcdb17bc40fa9e67581424c175daf410301f6f",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      790,
			createdBy: `COPY --chown=0:0 --chmod=644 /etc/envoy/envoy.yaml /etc/envoy/envoy.yaml # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:75a2c6152282d303c7c5fd2ffa8cad9c1ac0e4a0a909a72d147b69feba64deff",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      484,
			createdBy: `COPY --chown=0:0 --chmod=755 /docker-entrypoint.sh / # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:955a5dc17c1971170dae1b619dd415d9d3d481fb82bc45f2a4d1da2d4288791c",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      4679,
			createdBy: `COPY --chown=0:0 --chmod=755 /usr/local/bin/utils/su-exec /usr/local/bin/ # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:a94127188f8273b23ce07bdcb990b4ea3f0c145103578defc11ce1e477a744eb",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      36100758,
			createdBy: `COPY --chown=0:0 --chmod=755 /usr/local/bin/envoy /usr/local/bin/envoy # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:4f4fb700ef54461cfa02571ae0db9a0dc1e0cdb5577484a6d75e68dc38e8acc1",
			mediaType: api.MediaTypeOCIImageLayer,
			size:      32,
			createdBy: `COPY --chown=0:0 --chmod=755 /usr/local/bin/envoy.* /usr/local/bin/ # buildkit`,
		},
	},
}

var imageLinuxArm64 = image{
	url:      "http://test:5000/v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c",
	platform: "linux/arm64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:673aeee5c81c892477834e2b5e55575f16bfd52d9b841a1d8c524fb3805ee960",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      23703698,
			createdBy: `/bin/sh -c #(nop) ADD file:5f7cb4b44f843eaef6ae7ddb75dfc228a33d20cd974074ca23c1bb2cad7f77ad in / `,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:018b2790219d2003c0d437e634927887ee5cc3d8f985d7459adc5b2ff62d003f",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      851,
			createdBy: `/bin/sh -c set -xe 		&& echo '#!/bin/sh' > /usr/sbin/policy-rc.d 	&& echo 'exit 101' >> /usr/sbin/policy-rc.d 	&& chmod +x /usr/sbin/policy-rc.d 		&& dpkg-divert --local --rename --add /sbin/initctl 	&& cp -a /usr/sbin/policy-rc.d /sbin/initctl 	&& sed -i 's/^exit.*/exit 0/' /sbin/initctl 		&& echo 'force-unsafe-io' > /etc/dpkg/dpkg.cfg.d/docker-apt-speedup 		&& echo 'DPkg::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' > /etc/apt/apt.conf.d/docker-clean 	&& echo 'APT::Update::Post-Invoke { "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true"; };' >> /etc/apt/apt.conf.d/docker-clean 	&& echo 'Dir::Cache::pkgcache ""; Dir::Cache::srcpkgcache "";' >> /etc/apt/apt.conf.d/docker-clean 		&& echo 'Acquire::Languages "none";' > /etc/apt/apt.conf.d/docker-no-languages 		&& echo 'Acquire::GzipIndexes "true"; Acquire::CompressionTypes::Order:: "gz";' > /etc/apt/apt.conf.d/docker-gzip-indexes 		&& echo 'Apt::AutoRemove::SuggestsImportant "false";' > /etc/apt/apt.conf.d/docker-autoremove-suggests`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:509c77ce92ade89fbf09fe03b167023be51bf5a0c14c00487fa7a9ee33b55fc3",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      187,
			createdBy: `/bin/sh -c mkdir -p /run/systemd && echo 'docker' > /run/systemd/container`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:1cfa500dd01835df61b905a437de186592fa2adf6d6a3694a26c13f76c72b1f6",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      2617240,
			createdBy: `RUN |1 TARGETPLATFORM=linux/arm64 /bin/sh -c apt-get update && apt-get upgrade -y     && apt-get install --no-install-recommends -y ca-certificates     && apt-get autoremove -y && apt-get clean     && rm -rf /tmp/* /var/tmp/*     && rm -rf /var/lib/apt/lists/* # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:57227c32adb08b6f11b734f43a3c621a25a35833d2eaff6047612deffabea67f",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      120,
			createdBy: `RUN |1 TARGETPLATFORM=linux/arm64 /bin/sh -c mkdir -p /etc/envoy # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:97c59091ec632eb43a1f8ae51f48200b97a580b9fbf0c591ad5cccd12d2bd573",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      19994790,
			createdBy: `ADD linux/arm64/build_release_stripped/* /usr/local/bin/ # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:2a7ca8a5ead0b680d1e00675e8f0a3ee864e64173e7150fd056bd72659f69bd6",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      746,
			createdBy: `ADD configs/envoyproxy_io_proxy.yaml /etc/envoy/envoy.yaml # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:af66acd072fe6384d76fe0f86ccf256a9a6ae9c6cb8b2b38c9ea4241cb92aeca",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      3888,
			createdBy: `ADD linux/arm64/build_release/su-exec /usr/local/bin/ # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:f21ff7be3ac20eb86e923b81c6735b98f980e793bb88db26716944bb5f8730f0",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      1460,
			createdBy: `RUN |2 TARGETPLATFORM=linux/arm64 ENVOY_BINARY_SUFFIX=_stripped /bin/sh -c chown root:root /usr/local/bin/su-exec && adduser --group --system envoy # buildkit`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:68cf5c71735e492dc26366a69455c30b52e0787ebb8604909f77741f19883aeb",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      490,
			createdBy: `COPY ci/docker-entrypoint.sh / # buildkit`,
		},
	},
}

var imageWasmCompat = image{
	url:      "http://test:5000/v2/user/repo/manifests/sha256:03efb0078d32e24f3730afb13fc58b635bd4e9c6d5ab32b90af3922efc7f8672",
	platform: "linux/amd64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:d5e23ba78042fb166c603420339d92abb56a79bc8b689f4c84c96232a66be157",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      116164,
			createdBy: "COPY plugin.wasm ./ # buildkit",
		},
	},
}

var imageWindows = image{
	url:      "http://test:5000/v2/user/repo/manifests/v1.0",
	platform: "windows/amd64",
	filesystemLayers: []filesystemLayer{
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:47916aee02007e0e175e80deb2938cf8f95457b9abb555bd44dc461680dc552c",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      323887,
			createdBy: `cmd /S /C mkdir "C:\\Program\ Files\\envoy"`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:ba79ee4428b5ceec3026664126a146fd8c1041b478f3018ec0c90b78d7fe6355",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      331919,
			createdBy: `cmd /S /C setx path "%path%;c:\Program Files\envoy"`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:fd103a6c37aad8ffeaef6521612ed5a5153b104fffdb8bf3b6cf3d0beaaa49c4",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      12217107,
			createdBy: `cmd /S /C #(nop) ADD file:61df7bfb8255c0673d4ed25f961df5121141ee800202081e549fc36828624577 in C:\Program Files\envoy\ `,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:0fcfdc906e922391139a1c2d8f5d600066fa3b21c720a4024831471e2a8f0011",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      337530,
			createdBy: `cmd /S /C mkdir "C:\\ProgramData\\envoy"`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:f5ece8fbad694f5d1169c17ddd4217265cdf3dd886b71a8e9144f8b00e22de07",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      2410,
			createdBy: `cmd /S /C #(nop) ADD file:59ef68147ad4a3f10999e2e334cf60397fbcc6501b3949dd811afd7b8f03ca43 in C:\ProgramData\envoy\envoy.yaml `,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:8d3db7768af4371ec3f749f6816c8450687e276a883b8ca626a1fc1402fd32e0",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      419457,
			createdBy: `cmd /S /C powershell -Command "(cat C:\ProgramData\envoy\envoy.yaml -raw) -replace '/tmp/','C:\Windows\Temp\' | Set-Content -Encoding Ascii C:\ProgramData\envoy\envoy.yaml"`,
		},
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:9e17bb8cfb82c53b1793341a2dfb555e63088b1594d81d2b01106fae9a8aa60b",
			mediaType: api.MediaTypeDockerImageLayer,
			size:      1745,
			createdBy: `cmd /S /C #(nop) COPY file:4e78f00367722220f515590585490fc6d785cc05e3a59a54f965431fa3ef374e in C:\ `,
		},
	},
}

var imageTrivy = image{
	url:      "http://test:5000/v2/user/repo/manifests/v1.0",
	platform: "", // unknown
	filesystemLayers: []filesystemLayer{
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b",
			mediaType: api.MediaTypeModuleWasmImageLayer,
			size:      460018,
			fileName:  "wordpress.wasm",
		},
	},
}

// actual example is from ghcr.io/krustlet/oci-distribution/hello-wasm:v1
var imageKrustlet = image{
	url:      "http://test:5000/v2/user/repo/manifests/v1.0",
	platform: "", // unknown
	filesystemLayers: []filesystemLayer{
		{
			url:       "http://test:5000/v2/user/repo/blobs/sha256:f9c91f4c280ab92aff9eb03b279c4774a80b84428741ab20855d32004b2b983f",
			mediaType: api.MediaTypeWasmImageLayer,
			size:      1615998,
			fileName:  "module.wasm",
		},
	},
}

func TestNewImage(t *testing.T) {
	tests := []struct {
		name         string
		manifestJSON []byte
		configJSON   []byte
		expected     image
	}{
		{
			name:         "homebrew",
			manifestJSON: homebrew113VndOciImageManifestV1Json,
			configJSON:   homebrew113VndOciImageConfigV1Json,
			expected:     imageHomebrew,
		},
		{
			name:         "linux/amd64",
			manifestJSON: linuxAmd64VndOciImageManifestV1Json,
			configJSON:   linuxAmd64VndOciImageConfigV1Json,
			expected:     imageLinuxAmd64,
		},
		{
			name:         "wasm compat",
			manifestJSON: wasmCompatVndOciImageManifestV1Json,
			configJSON:   wasmCompatVndOciImageConfigV1Json,
			expected:     imageWasmCompat,
		},
		{
			name:         "windows",
			manifestJSON: windowsVndDockerImageManifestV1Json,
			configJSON:   windowsVndDockerImageConfigV1Json,
			expected:     imageWindows,
		},
		{
			name:         "trivy",
			manifestJSON: trivyVndOciImageManifestV1Json,
			configJSON:   []byte("{}"),
			expected:     imageTrivy,
		},
		{
			name:         "krustlet",
			manifestJSON: krustletVndOciImageManifestV1Json,
			configJSON:   []byte("{}"),
			expected:     imageKrustlet,
		},
	}

	const baseURL = "http://test:5000/v2/user/repo"
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var m imageManifestV1
			require.NoError(t, json.Unmarshal(tc.manifestJSON, &m))
			var c imageConfigV1
			require.NoError(t, json.Unmarshal(tc.configJSON, &c))
			m.URL = tc.expected.url
			require.Equal(t, tc.expected, newImage(baseURL, &m, &c))
		})
	}
}

func TestNewImage_FilesystemLayerAccessor(t *testing.T) {
	var image imageManifestV1
	require.NoError(t, json.Unmarshal(linuxArm64VndDockerImageManifestV1Json, &image))
	var config imageConfigV1
	require.NoError(t, json.Unmarshal(linuxArm64VndDockerImageConfigV1Json, &config))
	image.URL = "http://test:5000/v2/user/repo/manifests/sha256:f1cb90d4df0521842fe5f5c01a00032c76ba1743e1b2477589103373af06707c"

	for i := range imageLinuxArm64.filesystemLayers {
		require.Equal(t, &imageLinuxArm64.filesystemLayers[i], newImage("http://test:5000/v2/user/repo", &image, &config).FilesystemLayer(i))
	}
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
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, skipCreatedByPattern.MatchString(tc.createdBy))
		})
	}
}
