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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/httpclient"
)

func TestRoundTripper(t *testing.T) {
	type tagList struct {
		Name string
		Tags []string
	}
	expectedTagList := tagList{"homebrew/core/envoy", []string{"1.18.3", "1.18.3-1"}}

	u, err := url.Parse("https://ghcr.io/v2/homebrew/core/envoy/tags/list?n=100")
	require.NoError(t, err)

	ctx := httpclient.ContextWithTransport(context.Background(), &mock{t, fmt.Sprintf(`GET %s HTTP/1.1
Host: ghcr.io
User-Agent: Go-http-client/1.1
Authorization: Bearer QQ==

`, u.RequestURI()), expectedTagList})
	req := &http.Request{Method: "GET", URL: u, Header: http.Header{}}
	res, err := NewRoundTripper().RoundTrip(req.WithContext(ctx))
	require.NoError(t, err)
	res.Body.Close()
}

type mock struct {
	t            *testing.T
	request      string
	jsonResponse interface{}
}

func (m *mock) RoundTrip(req *http.Request) (*http.Response, error) {
	raw := new(bytes.Buffer)
	req.Write(raw) //nolint
	require.Equal(m.t, m.request, strings.ReplaceAll(raw.String(), "\r\n", "\n"))

	b, err := json.Marshal(m.jsonResponse)
	require.NoError(m.t, err)
	return &http.Response{Status: "200 OK", StatusCode: 200,
		Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(bytes.NewReader(b))}, nil
}
