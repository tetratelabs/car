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

package docker

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/tetratelabs/car/internal/httpclient"
)

// bearerAuth ensures there's a valid Bearer token prior to invoking the real request
type bearerAuth struct {
	token string
}

// NewRoundTripper creates an anonymous token for docker.io auth and re-uses it until it expires.
func NewRoundTripper() http.RoundTripper {
	return &bearerAuth{}
}

func (b *bearerAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	client := httpclient.New(httpclient.TransportFromContext(req.Context()))
	if b.token == "" {
		afterV2 := req.URL.Path[4:] // /v2/
		i := strings.Index(afterV2, "/manifests")
		if i == -1 {
			i = strings.Index(afterV2, "/blobs")
		}
		if i == -1 {
			return nil, fmt.Errorf("invalid docker.io URI: %s", req.RequestURI)
		}
		repository := afterV2[:i]
		token, err := b.newBearerToken(req.Context(), client, repository)
		if err != nil {
			return nil, err
		}
		b.token = token
	}

	// r2.cloudflarestorage.com doesn't like to see Authorization header in the request.
	// When the Authorization header is present, it returns 400 Bad Request.
	if !strings.HasSuffix(req.URL.Host, "r2.cloudflarestorage.com") {
		req.Header.Set("Authorization", "Bearer "+b.token)
	}
	req.Header.Set("User-Agent", "") // don't add implicit User-Agent
	return httpclient.TransportFromContext(req.Context()).RoundTrip(req)
}

// tokenResponse gets only the token as we don't run long enough to need refresh (>300s)
type tokenResponse struct {
	Token string `json:"token"`
}

func (b *bearerAuth) newBearerToken(ctx context.Context, client httpclient.HTTPClient, repository string) (string, error) {
	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repository)
	var tr tokenResponse
	if err := client.GetJSON(ctx, authURL, "application/json", &tr); err != nil {
		return "", err // wrapping doesn't help on this branch
	}
	if tr.Token == "" {
		return "", fmt.Errorf("invalid bearer token from %q", authURL)
	}
	return tr.Token, nil
}
