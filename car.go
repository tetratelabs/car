// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package car

import (
	"context"

	"github.com/tetratelabs/car/api"
	"github.com/tetratelabs/car/internal/reference"
	"github.com/tetratelabs/car/internal/registry"
)

// ParseReference is a simplified parser of OCI references that handle Docker
// familiar images. This is not strict, so a bad url will result in an HTTP
// error.
func ParseReference(ref string) (r api.Reference, err error) {
	return reference.Parse(ref)
}

// NewRegistry returns a new api.Registry appropriate for a Domain in an api.Reference.
func NewRegistry(ctx context.Context, refDomain string) (api.Registry, error) {
	return registry.New(ctx, refDomain)
}
