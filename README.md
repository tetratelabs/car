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
$ go build .
$ ./car -tvf envoyproxy/envoy:v1.18.3
-rwxr-xr-x	124804608	Jun 21 15:05:05	usr/local/bin/envoy
```
