// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package httpclient

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/test/httptest"
)

func TestGet_RetryDecisions(t *testing.T) {
	tests := []struct {
		name             string
		ctx              func(*testing.T) context.Context
		dialErr          error
		expectedErr      string
		expectedDials    int
		expectedRequests int
		expectedElapsed  time.Duration
	}{
		{
			name: "retries net error",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				return t.Context()
			},
			dialErr:          netError{err: errors.New("connection refused")},
			expectedDials:    2,
			expectedRequests: 1,
			expectedElapsed:  time.Second,
		},
		{
			name: "long deadline caps retry delay at 1s",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
				t.Cleanup(cancel)
				return ctx
			},
			dialErr:          netError{err: errors.New("connection refused")},
			expectedDials:    2,
			expectedRequests: 1,
			expectedElapsed:  time.Second,
		},
		{
			name: "cancel during retry sleep",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithCancel(t.Context())
				go func() {
					time.Sleep(500 * time.Millisecond)
					cancel()
				}()
				return ctx
			},
			dialErr:          netError{err: errors.New("connection refused")},
			expectedErr:      context.Canceled.Error(),
			expectedDials:    1,
			expectedRequests: 0,
			expectedElapsed:  500 * time.Millisecond,
		},
		{
			name: "does not retry non-net error",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				return t.Context()
			},
			dialErr:          errors.New("connection refused"),
			expectedErr:      `Get "$URL": connection refused`,
			expectedDials:    1,
			expectedRequests: 0,
		},
		{
			name: "does not dial canceled context",
			ctx: func(t *testing.T) context.Context {
				t.Helper()
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			expectedErr:      `Get "$URL": context canceled`,
			expectedDials:    0,
			expectedRequests: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				// Count requests that reach the server (past the dial stage).
				requests := 0
				ts := httptest.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					requests++
					w.WriteHeader(http.StatusOK)
				}))

				// Swap the client context with one that fails on the first dial.
				dials := 0
				client := ts.Client()
				transport := client.Transport.(*http.Transport)
				dialContext := transport.DialContext
				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					dials++
					if dials == 1 && tt.dialErr != nil {
						return nil, tt.dialErr
					}
					return dialContext(ctx, network, addr)
				}

				// Execute the request and verify the outcome.
				start := time.Now()
				res, err := Get(tt.ctx(t), client, ts.URL, http.Header{})
				if tt.expectedErr != "" {
					expectedErr := strings.ReplaceAll(tt.expectedErr, "$URL", ts.URL)
					require.EqualError(t, err, expectedErr)
				} else {
					require.NoError(t, err)
					defer res.Body.Close()
					require.Equal(t, http.StatusOK, res.StatusCode)
				}

				require.Equal(t, tt.expectedDials, dials)
				require.Equal(t, tt.expectedRequests, requests)
				require.Equal(t, tt.expectedElapsed, time.Since(start))
			})
		})
	}
}

func TestGet_Headers(t *testing.T) {
	tests := []struct {
		name           string
		header         http.Header
		expectedHeader http.Header
	}{
		{
			name: "sets custom headers",
			header: http.Header{
				"Accept":        {"application/json"},
				"Authorization": {"Bearer token123"},
			},
			expectedHeader: http.Header{
				"Accept":        {"application/json"},
				"Authorization": {"Bearer token123"},
			},
		},
		{
			name:           "clears user agent",
			header:         http.Header{"User-Agent": {""}},
			expectedHeader: http.Header{"User-Agent": {""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedHeader http.Header
			client := httptest.HTTPClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedHeader = r.Header
				w.WriteHeader(http.StatusOK)
			}))

			res, err := Get(t.Context(), client, "http://127.0.0.1:0/test", tt.header)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, http.StatusOK, res.StatusCode)
			require.Equal(t, tt.expectedHeader, capturedHeader)
		})
	}
}

func TestHttpClient_Get(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		header         http.Header
		expectedAccept []string
		expectedAuth   string
	}{
		{
			name:   "no headers",
			path:   "/repos/tetratelabs/getenvoy/releases",
			header: http.Header{},
		},
		{
			name: "Accept and Authorization",
			path: "/v2/tetratelabs/car-test/tags/list",
			header: http.Header{
				"Accept":        []string{"application/vnd.oci.image.index.v1+json"},
				"Authorization": []string{"Bearer QQ=="},
			},
			expectedAccept: []string{"application/vnd.oci.image.index.v1+json"},
			expectedAuth:   "Bearer QQ==",
		},
		{
			name: "multiple Accept values",
			path: "/v2/tetrate/car-test/manifests/latest",
			header: http.Header{
				"Accept": []string{
					"application/vnd.docker.distribution.manifest.list.v2+json",
					"application/vnd.docker.distribution.manifest.v2+json",
				},
				"Authorization": []string{"Bearer eyJhbGciOiJSUzI1NiIsInR5cC"},
			},
			expectedAccept: []string{
				"application/vnd.docker.distribution.manifest.list.v2+json",
				"application/vnd.docker.distribution.manifest.v2+json",
			},
			expectedAuth: "Bearer eyJhbGciOiJSUzI1NiIsInR5cC",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var capturedAccept []string
			var capturedAuth, capturedPath string
			ts := httptest.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedAccept = r.Header["Accept"]
				capturedAuth = r.Header.Get("Authorization")
				capturedPath = r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))

			client := New(ts.Client().Transport)
			_, _, err := client.Get(t.Context(), ts.URL+tc.path, tc.header)
			require.NoError(t, err)

			require.Equal(t, tc.expectedAccept, capturedAccept)
			require.Equal(t, tc.expectedAuth, capturedAuth)
			require.Equal(t, tc.path, capturedPath)
		})
	}
}

// TestHttpClient_Get_ErrorsOnBadRequest tests errors prior to the actual request
func TestHttpClient_Get_ErrorsOnBadRequest(t *testing.T) {
	client := New(http.DefaultTransport)
	_, _, err := client.Get(t.Context(), "https://api.github.com/\n", http.Header{})
	require.Error(t, err)
}

func TestHttpClient_Get_Body(t *testing.T) {
	tests := []struct {
		name              string
		contentType       string
		body              string
		expectedMediaType string
		expectedBody      string
	}{
		{
			name:              "returns body and media type",
			contentType:       "application/json",
			body:              `{"foo", "bar"}`,
			expectedMediaType: "application/json",
			expectedBody:      `{"foo", "bar"}`,
		},
		{
			name:              "strips content type parameters",
			contentType:       "application/json; charset=utf-8",
			expectedMediaType: "application/json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", tc.contentType)
				w.WriteHeader(http.StatusOK)
				if tc.body != "" {
					w.Write([]byte(tc.body))
				}
			}))

			client := New(ts.Client().Transport)
			body, mediaType, err := client.Get(t.Context(), ts.URL, http.Header{})
			require.NoError(t, err)
			defer body.Close()

			require.Equal(t, tc.expectedMediaType, mediaType)
			if tc.expectedBody != "" {
				b, err := io.ReadAll(body)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, string(b))
			}
		})
	}
}

func TestTransportFromContext(t *testing.T) {
	require.Equal(t, http.DefaultTransport, TransportFromContext(t.Context()))

	var r http.RoundTripper = &http.Transport{}
	ctx := ContextWithTransport(t.Context(), r)
	require.Same(t, r, TransportFromContext(ctx))
}

var _ net.Error = netError{}

type netError struct {
	err error
}

func (e netError) Error() string {
	return e.err.Error()
}

func (e netError) Unwrap() error {
	return e.err
}

func (netError) Timeout() bool {
	return false
}

func (netError) Temporary() bool {
	return false
}
