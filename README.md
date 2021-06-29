[![Build](https://github.com/tetratelabs/car/workflows/build/badge.svg)](https://github.com/tetratelabs/car)
[![Coverage](https://codecov.io/gh/tetratelabs/car/branch/master/graph/badge.svg)](https://codecov.io/gh/tetratelabs/car)
[![Go Report Card](https://goreportcard.com/badge/github.com/tetratelabs/car)](https://goreportcard.com/report/github.com/tetratelabs/car)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

# car

`car` is like `tar`, but for containers!

Mainly, `car` lets you list or extract files from an OCI (possibly Docker) image, regardless of the platform it was
built for. For example, you can extract files from a `windows/amd64` image even if you are running `linux/arm64`.

## Experimental!

Right now, only list works:

```bash
$ go build -o car main.go
$ ./car --platform linux/amd64 -tvvf alpine:3.14.0
https://index.docker.io/v2/library/alpine/manifests/sha256:1775bebec23e1f3ce486989bfc9ff3c4e951690df84aa9f926497d82f2ffca9d platform=linux/amd64 totalLayerSize: 2811478
https://index.docker.io/v2/library/alpine/blobs/sha256:5843afab387455b37944e709ee8c78d7520df80f8d01cf7f861aae63beeddb6b size=2811478 CreatedBy: /bin/sh -c #(nop) ADD file:f278386b0cef68136129f5f58c52445590a417b624d62bca158d4dc926c340df in / 
-rwxr-xr-x	829000	Jun 14 18:24:54	bin/busybox
-rw-r--r--	7	Jun 15 22:32:26	etc/alpine-release
-rw-r--r--	7	Jun 15 22:34:40	etc/apk/arch
--snip--
```
