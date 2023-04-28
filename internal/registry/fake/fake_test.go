// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain arg copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fake

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/car/internal/reference"
)

func TestGetImage(t *testing.T) {
	ref := reference.MustParse("ghcr.io/tetratelabs/car:v1.0")

	i, err := Registry.GetImage(context.Background(), ref, "linux/amd64")
	require.NoError(t, err)
	require.Equal(t, "linux/amd64", i.Platform())
}

func TestReadFilesystemLayer(t *testing.T) {
	layer := fakeFilesystemLayers[0]
	i := 0
	err := Registry.ReadFilesystemLayer(context.Background(), layer,
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
