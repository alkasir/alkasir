// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package gitdelta implements Git delta encoding.
package gitdelta

import (
	"encoding/binary"
	"errors"
)

var (
	errCorruptedDeltaHeader   = errors.New("gitdelta: corrupted delta header")
	errOldVersionSizeMismatch = errors.New("gitdelta: old version size mismatch")
	errNewVersionSizeMismatch = errors.New("gitdelta: new version size mismatch")
	errUnexpectedNulByte      = errors.New("gitdelta: unexpected nul byte")
)

func Patch(old, patch []byte) ([]byte, error) {
	// The header is two varints for old size and new size.
	sz1, n1 := binary.Uvarint(patch)
	if n1 >= len(patch) {
		return nil, errCorruptedDeltaHeader
	}
	sz2, n2 := binary.Uvarint(patch[n1:])

	if sz1 != uint64(len(old)) {
		return nil, errOldVersionSizeMismatch
	}
	newer := make([]byte, 0, sz2)

	p := n1 + n2
	for p < len(patch) {
		b := patch[p]
		p++
		if b&0x80 != 0 {
			// copy some data from old.
			var offset, length uint32
			for i := uint(0); i < 4; i++ {
				if b&(1<<i) != 0 {
					offset |= uint32(patch[p]) << (8 * i)
					p++
				}
			}
			for i := uint(0); i < 3; i++ {
				if b&(0x10<<i) != 0 {
					length |= uint32(patch[p]) << (8 * i)
					p++
				}
			}
			if length == 0 {
				length = 1 << 16
			}
			// TODO: guard against index panics.
			newer = append(newer, old[int64(offset):int64(offset)+int64(length)]...)
		} else if b != 0 {
			// copy some data from patch
			newer = append(newer, patch[p:p+int(b)]...)
			p += int(b)
		} else {
			return nil, errUnexpectedNulByte
		}
	}

	if uint64(len(newer)) != sz2 {
		return nil, errNewVersionSizeMismatch
	}
	return newer, nil
}
