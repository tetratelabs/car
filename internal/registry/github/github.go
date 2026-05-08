// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"net/http"

	"github.com/tetratelabs/car/internal/httpclient"
)

type fixedBearerToken struct{}

// NewRoundTripper creates re-uses a fake bearer token on each request.
func NewRoundTripper() http.RoundTripper {
	return &fixedBearerToken{}
}

func (f *fixedBearerToken) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer QQ==")
	client := &http.Client{Transport: httpclient.TransportFromContext(req.Context())}
	return httpclient.Get(req.Context(), client, req.URL.String(), req.Header)
}
