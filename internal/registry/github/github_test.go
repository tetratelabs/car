// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/httpclient"
	"github.com/tetratelabs/car/internal/test/httptest"
)

func TestRoundTripper(t *testing.T) {
	type tagList struct {
		Name string
		Tags []string
	}
	expectedTagList := tagList{"homebrew/core/envoy", []string{"1.18.3", "1.18.3-1"}}

	var actualAuth string
	client := httptest.HTTPClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedTagList)
	}))
	ctx := httpclient.ContextWithTransport(t.Context(), client.Transport)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://ghcr.io/v2/homebrew/core/envoy/tags/list?n=100", http.NoBody)
	require.NoError(t, err)
	res, err := NewRoundTripper().RoundTrip(req)
	require.NoError(t, err)
	res.Body.Close()

	require.Equal(t, "Bearer QQ==", actualAuth)
}
