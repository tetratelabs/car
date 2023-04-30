// Copyright 2023 Tetrate
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

package reference

import (
	"errors"
	"strings"

	"github.com/tetratelabs/car/internal"
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
func Parse(ref string) (r *Reference, err error) {
	if ref == "" {
		err = errors.New("invalid reference format")
		return
	}

	// First, check to see if there's at least one colon. If not, this cannot
	// be a tagged image.
	indexColon := strings.LastIndexByte(ref, byte(':'))
	indexSlash := strings.IndexByte(ref, byte('/'))
	if indexColon == -1 || indexSlash > indexColon /* e.g. host:80/image */ {
		err = errors.New("expected tagged reference")
		return

	}

	r = &Reference{}
	r.tag = ref[indexColon+1:]
	remaining := ref[0:indexColon]

	// See if this is a familiar official docker image. e.g. "alpine:3.14.0"
	if indexSlash == -1 {
		r.domain = "docker.io"
		r.path = "library/" + remaining
		return
	}

	// See if this is an official docker image. e.g. "envoyproxy/envoy:v1.18.3"
	if strings.LastIndexByte(ref, byte('/')) == indexSlash &&
		strings.IndexByte(remaining, byte('.')) == -1 {
		r.domain = "docker.io"
		r.path = remaining
		return
	}

	// Otherwise, the part leading to the first slash is the domain.
	r.domain = remaining[0:indexSlash]
	r.path = remaining[indexSlash+1:]
	return
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
