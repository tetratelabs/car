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

package github

import (
	"net/http"

	"github.com/tetratelabs/car/internal/httpclient"
)

type fixedBearerToken struct {}

// NewRoundTripper creates re-uses a fake bearer token on each request.
func NewRoundTripper() http.RoundTripper {
	return &fixedBearerToken{}
}

func (f *fixedBearerToken) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer QQ==")
	return httpclient.TransportFromContext(req.Context()).RoundTrip(req)
}
