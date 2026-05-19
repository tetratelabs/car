// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/reference"
)

func TestGetImage(t *testing.T) {
	ref := reference.MustParse("ghcr.io/tetratelabs/car:v1.0")

	i, err := Registry.GetImage(t.Context(), ref, "linux/amd64")
	require.NoError(t, err)
	require.Equal(t, "linux/amd64", i.Platform())
}

func TestReadFilesystemLayer(t *testing.T) {
	layer := &fakeFilesystemLayers[0]
	i := 0
	err := Registry.ReadFilesystemLayer(t.Context(), layer,
		func(name string, size int64, mode os.FileMode, modTime time.Time, reader io.Reader) error {
			require.Equal(t, fakeFiles[0][i].name, name)
			require.Equal(t, fakeFiles[0][i].size, size)
			require.Equal(t, fakeFiles[0][i].mode, mode)
			require.Equal(t, fakeFiles[0][i].modTimeRFC3339, modTime.Format(time.RFC3339))

			// verify the fake body exists
			b, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.Equal(t, fakeFiles[0][i].size, int64(len(b)))

			i++
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, len(fakeFiles[0]), i)
}
