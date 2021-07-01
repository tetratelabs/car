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

package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	urlpkg "net/url"
)

// HTTPClient is a convenience wrapper for http.Client that consolidates common logic.
type HTTPClient interface {
	// Get returns the body and media type of the URL using the provided context. The caller must close the body.
	//
	// This is optimized for easy content negotiation. Hence, the returned mediaType is stripped of qualifiers.
	// Ex. "Content-Type: application/json; charset=utf-8" will return mediaType "application/json"
	Get(ctx context.Context, url string, header http.Header) (body io.ReadCloser, mediaType string, err error)
	// GetJSON is a convenience function that calls json.Unmarshal after Get.
	GetJSON(ctx context.Context, url string, accept string, v interface{}) error
}

type httpClient struct{ client http.Client }

// New returns a client that implicitly authenticates when it needs to
// Use ContextWithTransport when testing.
func New(transport http.RoundTripper) HTTPClient {
	return &httpClient{client: http.Client{Transport: transport}}
}

type contextClientTransportKey struct{}

// TransportFromContext returns an http.RoundTripper for use as http.Client Transport from the context or nil
func TransportFromContext(ctx context.Context) http.RoundTripper {
	if v, ok := ctx.Value(contextClientTransportKey{}).(http.RoundTripper); ok {
		return v
	}
	return http.DefaultTransport
}

// ContextWithTransport returns a context with a http.RoundTripper for use as http.Client Transport
func ContextWithTransport(ctx context.Context, transport http.RoundTripper) context.Context {
	return context.WithValue(ctx, contextClientTransportKey{}, transport)
}

func (h *httpClient) Get(ctx context.Context, url string, header http.Header) (io.ReadCloser, string, error) {
	u, err := urlpkg.Parse(url)
	if err != nil {
		return nil, "", err
	}

	hdr := http.Header{}
	if len(header) > 0 {
		hdr = header.Clone()
	}
	hdr.Set("User-Agent", "") // don't add implicit User-Agent
	req := &http.Request{Method: http.MethodGet, URL: u, Header: hdr}
	res, err := h.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, "", err
	}

	if res.StatusCode != http.StatusOK {
		res.Body.Close() //nolint
		return nil, "", fmt.Errorf("received %v status code from %q", res.StatusCode, url)
	}

	contentType := res.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(contentType) // strip qualifiers
	return res.Body, mediaType, nil
}

func (h *httpClient) GetJSON(ctx context.Context, url, accept string, v interface{}) error {
	header := http.Header{}
	header.Add("Accept", accept)
	body, _, err := h.Get(ctx, url, header)
	if err != nil {
		return err // wrapping doesn't help on this branch
	}
	defer body.Close()         //nolint
	b, err := io.ReadAll(body) // fully read the response
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("error unmarshalling %v: %w", v, err)
	}
	return nil
}
