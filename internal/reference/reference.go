// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package reference

import (
	"errors"
	"strings"

	"github.com/tetratelabs/car/internal"
)

const (
	dockerHubHost      = "docker.io"
	dockerHubIndexHost = "index.docker.io"
)

type Reference struct {
	internal.CarOnly

	domain, path, tag string
}

// MustParse calls Parse or panics on error.
func MustParse(ref string) *Reference {
	r, err := Parse(ref)
	if err != nil {
		panic(err)
	}
	return r
}

// Parse is a simplified parser of OCI references that handle Docker
// familiar images. This is not strict, so a bad url will result in an HTTP
// error.
func Parse(ref string) (*Reference, error) {
	if ref == "" {
		return nil, errors.New("invalid reference format")
	}

	// First, check to see if there's at least one colon. If not, this cannot
	// be a tagged image.
	indexColon := strings.LastIndexByte(ref, byte(':'))
	indexSlash := strings.IndexByte(ref, byte('/'))
	if indexColon == -1 || indexSlash > indexColon /* e.g. host:80/image */ {
		return nil, errors.New("expected tagged reference")
	}

	r := &Reference{}
	r.tag = ref[indexColon+1:]
	remaining := ref[0:indexColon]

	// See if this is a familiar official docker image. e.g. "alpine:3.14.0"
	if indexSlash == -1 {
		r.domain = dockerHubIndexHost
		r.path = "library/" + remaining
		return r, nil
	}

	// See if this is an official docker image. e.g. "envoyproxy/envoy:v1.18.3"
	if strings.LastIndexByte(ref, byte('/')) == indexSlash &&
		strings.IndexByte(remaining, byte('.')) == -1 {
		r.domain = dockerHubIndexHost
		r.path = remaining
		return r, nil
	}

	// Otherwise, the part leading to the first slash is the domain.
	r.domain = remaining[0:indexSlash]

	// Finally, any direct reference to docker.io should use the index
	if r.domain == dockerHubHost {
		r.domain = dockerHubIndexHost
	}

	r.path = remaining[indexSlash+1:]
	return r, nil
}

func (r *Reference) Domain() string {
	return r.domain
}

func (r *Reference) Path() string {
	return r.path
}

func (r *Reference) Tag() string {
	return r.tag
}

// String implements fmt.Stringer
func (r *Reference) String() string {
	return r.domain + "/" + r.path + "/" + r.tag
}
