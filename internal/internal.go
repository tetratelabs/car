// Copyright car contributors
// SPDX-License-Identifier: Apache-2.0

package internal

// CarOnly indicates by type the implementation is internal only.
//
// Design credit to https://github.com/tetratelabs/wazero/pull/1396
type CarOnly interface {
	carOnly()
}

type CarOnlyType struct{}

func (CarOnlyType) carOnly() {}
