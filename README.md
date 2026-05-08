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
$ ./car --platform linux/amd64 -tf envoyproxy/envoy:v1.38.0 'Files/Program Files/envoy/envoy.exe'
error: Files/Program Files/envoy/envoy.exe not found in layer

# extract a file from an image
$ ./car --platform linux/amd64 --strip-components 3 --created-by-pattern 'COPY amd64/envoy /usr/local/bin/envoy' -xvvf istio/proxyv2:1.28.6 usr/local/bin/envoy && file envoy
https://index.docker.io/v2/istio/proxyv2/manifests/sha256:b6a8f710bce3368d6c7ad2e67fa169b9eebaa999423c7ad8b34cb3ea36befee0 platform=linux/amd64 totalLayerSize: 99843393
https://index.docker.io/v2/istio/proxyv2/blobs/sha256:b6a20e28a0b487a584382c862c3fbf6f8c7bad226d30b9886860622cadd307d8 size=41026126
CreatedBy: COPY amd64/envoy /usr/local/bin/envoy # buildkit
-rwxr-xr-x	138699504	Apr 10 21:05:48	usr/local/bin/envoy
envoy: ELF 64-bit LSB pie executable, x86-64, version 1 (SYSV), dynamically linked, interpreter /lib64/ld-linux-x86-64.so.2, for GNU/Linux 3.2.0, not stripped

# try a platform you may not usually be able to poke
$ ./car -tvvf chocolateyfest/chocolatey:latest
https://index.docker.io/v2/chocolateyfest/chocolatey/manifests/latest platform=windows/amd64 totalLayerSize: 24102006
https://index.docker.io/v2/chocolateyfest/chocolatey/blobs/sha256:6d2d8da2960b0044c22730be087e6d7b197ab215d78f9090a3dff8cb7c40c241 size=24102006
CreatedBy: cmd /S /C powershell iex(iwr -useb https://chocolatey.org/install.ps1)
-rw-r--r--	44245	May  5 03:09:14	Files/ProgramData/chocolatey/CREDITS.txt
-rw-r--r--	670	May  5 03:09:14	Files/ProgramData/chocolatey/LICENSE.txt
-rw-r--r--	2283	May  5 03:09:14	Files/ProgramData/chocolatey/bin/RefreshEnv.cmd
--snip--

# try a multi-platform image
$ ./car -tvvf alpine:3.23.4
error: choose a platform: linux/386, linux/amd64, linux/arm, linux/arm64, linux/ppc64le, linux/riscv64, linux/s390x, unknown/unknown
$ ./car --platform linux/arm64 -tvvf alpine:3.23.4
https://index.docker.io/v2/library/alpine/manifests/sha256:378c4c5418f7493bd500ad21ffb43818d0689daaad43e3261859fb417d1481a0 platform=linux/arm64 totalLayerSize: 4199870
https://index.docker.io/v2/library/alpine/blobs/sha256:d17f077ada118cc762df373ff803592abf2dfa3ddafaa7381e364dd27a88fca7 size=4199870
CreatedBy: ADD alpine-minirootfs-3.23.4-aarch64.tar.gz / # buildkit
-rwxr-xr-x	919304	Dec 16 23:19:28	bin/busybox
-rw-r--r--	7	Apr 15 13:50:36	etc/alpine-release
--snip--

# try a container image that contains a wasm file
$ ./car --platform linux/amd64 -tvvf ghcr.io/corazawaf/coraza-proxy-wasm:0.6.0
https://ghcr.io/v2/corazawaf/coraza-proxy-wasm/manifests/sha256:65d6009b9da2e8965e592a08b74a86725435fc01aa39c756dce0bd5ea64b3f4e platform=linux/amd64 totalLayerSize: 4830982
https://ghcr.io/v2/corazawaf/coraza-proxy-wasm/blobs/sha256:5e0a42d17c1b3b8a6ded98ef90e1a1a6c2dd49729e6ca1fa311ba0372e6f2952 size=4830982
CreatedBy: COPY build/main.wasm /plugin.wasm # buildkit
-rw-r--r--	18565738	Jul  7 17:46:53	plugin.wasm
```
