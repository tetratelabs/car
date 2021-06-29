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
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	httpclient "github.com/tetratelabs/car/internal/httpclient"
)

type rfc3339Nano struct {
	time.Time
}

type tokenResponse struct {
	Token     string      `json:"token"`
	ExpiresIn int         `json:"expires_in"`
	IssuedAt  rfc3339Nano `json:"issued_at"`
}

func newTokenResponse(token string, expiresIn int, issuedAt time.Time) *tokenResponse {
	return &tokenResponse{token, expiresIn, rfc3339Nano{issuedAt}}
}

func (r *rfc3339Nano) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Format(time.RFC3339Nano))
}

func (r *rfc3339Nano) UnmarshalJSON(b []byte) (err error) {
	s := string(b[1 : len(b)-1]) // without quotes
	t, err := time.Parse(time.RFC3339Nano, s)
	r.Time = t
	return
}

// bearerAuth ensures there's a valid Bearer token prior to invoking the real request
type bearerAuth struct {
	repository     string
	token          string
	tokenExpiresAt time.Time
}

// NewRoundTripper creates an anonymous token for docker.io auth and re-uses it until it expires.
func NewRoundTripper(repository string) http.RoundTripper {
	return &bearerAuth{repository: repository}
}

func (b *bearerAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	client := httpclient.New(httpclient.TransportFromContext(req.Context()))
	if b.token == "" || time.Now().After(b.tokenExpiresAt) {
		if err := b.refreshBearerToken(req.Context(), client); err != nil {
			return nil, err
		}
	}

	req.Header.Set("Authorization", "Bearer "+b.token)
	return httpclient.TransportFromContext(req.Context()).RoundTrip(req)
}

func (b *bearerAuth) refreshBearerToken(ctx context.Context, client httpclient.HTTPClient) error {
	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", b.repository)
	var tr tokenResponse
	if err := client.GetJSON(ctx, authURL, "application/json", &tr); err != nil {
		return err // wrapping doesn't help on this branch
	}

	b.token = tr.Token
	b.tokenExpiresAt = tr.IssuedAt.Add(time.Duration(tr.ExpiresIn) * time.Second)

	if b.token == "" || time.Now().After(b.tokenExpiresAt) {
		return fmt.Errorf("invalid bearer token from %q", authURL)
	}
	return nil
}
