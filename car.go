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
