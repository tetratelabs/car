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

package internal

const (
	// ArchAmd64 is a Platform.Architecture a.k.a. "x86_64"
	ArchAmd64 = "amd64"
	// ArchArm64 is a Platform.Architecture a.k.a. "aarch64"
	ArchArm64 = "arm64"
	// OSDarwin is a Platform.OS a.k.a. "macOS"
	OSDarwin = "darwin"
	// OSLinux is a Platform.OS
	OSLinux = "linux"
	// OSWindows is a Platform.OS
	OSWindows = "windows"
)

// IsValidArch returns true on a supported runtime.GOARCH
func IsValidArch(arch string) bool {
	return arch == ArchAmd64 || arch == ArchArm64
}

// IsValidOS returns true on a supported runtime.GOOS
func IsValidOS(os string) bool {
	return os == OSDarwin || os == OSLinux || os == OSWindows
}
