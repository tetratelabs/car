// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	neturl "net/url"
	"time"
)

// HTTPClient is a convenience wrapper for http.Client that consolidates common logic.
type HTTPClient interface {
	// Get returns the body and media type of the URL using the provided context. The caller must close the body.
	//
	// This is optimized for easy content negotiation. Hence, the returned mediaType is stripped of qualifiers.
	// e.g. "Content-Type: application/json; charset=utf-8" will return mediaType "application/json"
	Get(ctx context.Context, url string, header http.Header) (body io.ReadCloser, mediaType string, err error)

	// GetJSON is a convenience function that calls json.Unmarshal after Get.
	GetJSON(ctx context.Context, url string, accept string, v any) error
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
	res, err := Get(ctx, &h.client, url, header)
	if err != nil {
		return nil, "", err
	}

	if res.StatusCode != http.StatusOK {
		res.Body.Close() //nolint:errcheck,gosec // error on close is unactionable for failed response
		return nil, "", fmt.Errorf("received %v status code from %q", res.StatusCode, url)
	}

	contentType := res.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType) // strip qualifiers
	if err != nil {
		mediaType = contentType
	}
	return res.Body, mediaType, nil
}

// Get GETs rawURL with the given headers and one retry on transient network error.
func Get(ctx context.Context, client *http.Client, rawURL string, header http.Header) (*http.Response, error) {
	get := func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
		if err != nil {
			return nil, err
		}
		req.Header = header.Clone()
		return client.Do(req)
	}

	resp, err := get()

	// Return unless this hit a transient network error worth retrying.
	if resp != nil || err == nil || ctx.Err() != nil || !isNetError(err) {
		return resp, err
	}

	// Wait up to 1s before retrying, or bail if the context is canceled.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(time.Second):
	}

	return get()
}

// isNetError unwraps url.Error so transient dial/TLS failures are retried
// while HTTP-level errors (4xx, 5xx) are not.
func isNetError(err error) bool {
	if urlErr, ok := errors.AsType[*neturl.Error](err); ok {
		err = urlErr.Err
	}

	netErr, ok := errors.AsType[net.Error](err)
	return ok && netErr != nil
}

func (h *httpClient) GetJSON(ctx context.Context, url, accept string, v any) error {
	header := http.Header{}
	header.Add("Accept", accept)
	body, _, err := h.Get(ctx, url, header)
	if err != nil {
		return err // wrapping doesn't help on this branch
	}
	defer body.Close()         //nolint:errcheck // error on close is unactionable
	b, err := io.ReadAll(body) // fully read the response
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("error unmarshalling %v: %w", v, err)
	}
	return nil
}
