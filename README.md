[![Build](https://github.com/tetratelabs/car/workflows/build/badge.svg)](https://github.com/tetratelabs/car)
[![Coverage](https://codecov.io/gh/tetratelabs/car/branch/master/graph/badge.svg)](https://codecov.io/gh/tetratelabs/car)
[![Go Report Card](https://goreportcard.com/badge/github.com/tetratelabs/car)](https://goreportcard.com/report/github.com/tetratelabs/car)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

# car

`car` is like `tar`, but for containers!

Mainly, `car` lets you list or extract files from an OCI (possibly Docker) image, regardless of the platform it was
built for. For example, you can extract files from a `windows/amd64` image even if you are running `linux/arm64`.

## Experimental!

```bash
$ go build ./cmd/car

# verify a file you think is in an image is really there
$ ./car -tf envoyproxy/envoy-alpine:v1.18.3 'Files/Program Files/envoy/envoy.exe'
error: Files/Program Files/envoy/envoy.exe not found in layer
$ ./car -tf envoyproxy/envoy-windows:v1.18.3 'Files/Program Files/envoy/envoy.exe'
Files/Program Files/envoy/envoy.exe

# extract a file from an image
$ ./car --strip-components 3 --created-by-pattern 'COPY envoy /usr/local/bin/envoy' -xvvf istio/proxyv2:1.10.3 && test -f envoy
https://index.docker.io/v2/istio/proxyv2/manifests/1.10.3 platform=linux/amd64 totalLayerSize: 95073366
https://index.docker.io/v2/istio/proxyv2/blobs/sha256:5afc65eb63c65ce691cc003c8b26820b7d984181b4871a2735e92cbf69595671 size=26407160
CreatedBy: COPY envoy /usr/local/bin/envoy # buildkit
-rwxr-xr-x	100920696	Jul 15 14:15:57	usr/local/bin/envoy

# try a platform you may no usually be able to poke
$ ./car -tvvf chocolateyfest/chocolatey:latest
https://index.docker.io/v2/chocolateyfest/chocolatey/manifests/latest platform=windows/amd64 totalLayerSize: 24102006
https://index.docker.io/v2/chocolateyfest/chocolatey/blobs/sha256:6d2d8da2960b0044c22730be087e6d7b197ab215d78f9090a3dff8cb7c40c241 size=24102006
CreatedBy: cmd /S /C powershell iex(iwr -useb https://chocolatey.org/install.ps1)
-rw-r--r--	44245	May  5 02:09:14	Files/ProgramData/chocolatey/CREDITS.txt
-rw-r--r--	670	May  5 02:09:14	Files/ProgramData/chocolatey/LICENSE.txt
-rw-r--r--	2283	May  5 02:09:14	Files/ProgramData/chocolatey/bin/RefreshEnv.cmd
--snip--

# try a multi-platform image
$ ./car -tvvf alpine:3.14.0
error: choose a platform: linux/386, linux/amd64, linux/arm, linux/arm64, linux/ppc64le, linux/s390x
$ ./car --platform linux/arm64 -tvvf alpine:3.14.0
https://index.docker.io/v2/library/alpine/manifests/sha256:53b74ddfc6225e3c8cc84d7985d0f34666e4e8b0b6892a9b2ad1f7516bc21b54 platform=linux/arm64 totalLayerSize: 2709626
https://index.docker.io/v2/library/alpine/blobs/sha256:58ab47519297212468320b23b8100fc1b2b96e8d342040806ae509a778a0a07a size=2709626
CreatedBy: /bin/sh -c #(nop) ADD file:6797caacbfe41bfe44000b39ed017016c6fcc492b3d6557cdaba88536df6c876 in /
-rwxr-xr-x	878176	Jun 14 18:24:54	bin/busybox
-rw-r--r--	7	Jun 15 22:32:26	etc/alpine-release
--snip--

# try a wasm image
$ ./car -tvvf ghcr.io/aquasecurity/trivy-module-wordpress:latest
https://ghcr.io/v2/aquasecurity/trivy-module-wordpress/manifests/latest platform= totalLayerSize: 460018
https://ghcr.io/v2/aquasecurity/trivy-module-wordpress/blobs/sha256:3daa3dac086bd443acce56ffceb906993b50c5838b4489af4cd2f1e2f13af03b size=460018
CreatedBy:
-rw-r--r--	460018	Apr 25 08:22:32	wordpress.wasm
```
