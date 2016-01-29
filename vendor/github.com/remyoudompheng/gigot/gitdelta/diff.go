// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitdelta

import (
	"bytes"
	"encoding/binary"
)

// This file uses Rabin hashing to delta encode.

// hashChunks hashes chunks input[k*W+1 : (k+1)*W]
// and returns a map from hashes to index in input buffer.
func hashChunks(input []byte) hashmap {
	nbHash := len(input) / _W
	if nbHash >= 0x7fffffff {
		panic("input too large")
	}
	var hashes hashmap
	hashes.Init(int32(nbHash))
	for i := (nbHash - 1) * _W; i > 0; i -= _W {
		// on collision overwrite with smallest index.
		h := hashRabin(input[i : i+_W])
		hashes.Set(h, i)
	}
	return hashes
}

// Diff computes a delta from data1 to data2. The
// result is such that Patch(data1, Diff(data1, data2)) == data2.
func Diff(data1, data2 []byte) []byte {
	// Store lengths of inputs.
	patch := make([]byte, 32)
	n1 := binary.PutUvarint(patch, uint64(len(data1)))
	n2 := binary.PutUvarint(patch[n1:], uint64(len(data2)))
	patch = patch[:n1+n2]

	// First hash chunks of data1.
	hashes := hashChunks(data1)

	// Compute rolling hashes of data2 and see whether
	// we recognize parts of data1.
	var p uint32
	lastmatch := -1
	for i := 0; i < len(data2); i++ {
		b := data2[i]
		if i < _W {
			p = (p << 8) ^ uint32(b) ^ _T[uint8(p>>(degree-8))]
			continue
		}
		// Invariant: i >= W and p == hashRabin(data2[i-W:i])
		//if p != hashRabin(data2[i-_W:i]) {
		//	println(p, hashRabin(data2[i-_W:i]))
		//	panic("p != hashRabin(data2[i-_W:i])")
		//}

		refi, ok := hashes.Get(p)
		if ok && bytes.Equal(data1[refi:refi+_W], data2[i-_W:i]) {
			// We have a match! Try to extend it left and right.
			testi := i - _W
			for refi > 0 && testi > lastmatch+1 && data1[refi-1] == data2[testi-1] {
				refi--
				testi--
			}
			refj, testj := refi+i-testi, i
			for refj < len(data1) && testj < len(data2) && data1[refj] == data2[testj] {
				refj++
				testj++
			}

			// Now data1[refi:refj] == data2[testi:testj]
			patch = appendInlineData(patch, data2[lastmatch+1:testi])
			patch = appendRefData(patch, uint32(refi), uint32(refj-refi))

			// Skip bytes and update hash.
			i = testj + _W - 1
			lastmatch = testj - 1
			if i >= len(data2) {
				break
			}
			p = hashRabin(data2[testj : testj+_W])
			continue
		}
		// Cancel out data2[i-W] and take data2[i]
		p ^= _U[data2[i-_W]]
		p = (p << 8) ^ uint32(b) ^ _T[uint8(p>>(degree-8))]
	}
	patch = appendInlineData(patch, data2[lastmatch+1:])
	return patch
}

// appendInlineData encodes inline data in a patch.
func appendInlineData(patch, data []byte) []byte {
	for len(data) > 0x7f {
		patch = append(patch, 0x7f)
		patch = append(patch, data[:0x7f]...)
		data = data[0x7f:]
	}
	if len(data) > 0 {
		patch = append(patch, byte(len(data)))
		patch = append(patch, data...)
	}
	return patch
}

// appendRefData encodes reference to original data in a delta.
func appendRefData(patch []byte, off, length uint32) []byte {
	for length > 1<<16 {
		// emit opcode for length 1<<16.
		switch {
		case off>>8 == 0:
			patch = append(patch, 0x81, byte(off))
		case off>>16 == 0:
			patch = append(patch, 0x83, byte(off), byte(off>>8))
		case off>>24 == 0:
			patch = append(patch, 0x87,
				byte(off), byte(off>>8), byte(off>>16))
		default:
			patch = append(patch, 0x8f,
				byte(off), byte(off>>8), byte(off>>16), byte(off>>24))
		}
		off += 1 << 16
		length -= 1 << 16
	}

	iop := len(patch)
	patch = append(patch, 0)
	op := byte(0x80)

	if b := byte(off); b != 0 {
		op |= 1
		patch = append(patch, b)
	}
	if b := byte(off >> 8); b != 0 {
		op |= 2
		patch = append(patch, b)
	}
	if b := byte(off >> 16); b != 0 {
		op |= 4
		patch = append(patch, b)
	}
	if b := byte(off >> 24); b != 0 {
		op |= 8
		patch = append(patch, b)
	}

	if b := byte(length); b != 0 {
		op |= 0x10
		patch = append(patch, b)
	}
	if b := byte(length >> 8); b != 0 {
		op |= 0x20
		patch = append(patch, b)
	}

	patch[iop] = op
	return patch
}

// A hashmap provides similar functionality as a map[uint32]int.
type hashmap struct {
	Bits    uint
	Map     []int32 // lowbits -> index in hashEntry
	Entries []hashEntry
}

type hashEntry struct {
	Next int32 // index in hashmap.Entries, we don't support >4GB blobs.
	Key  uint32
	Val  int
}

func (h *hashmap) Init(size int32) {
	if size < 0 {
		panic("negative size")
	}
	bits := uint(4)
	for uint32(1)<<bits < uint32(size) && bits < 31 {
		bits++
	}
	h.Bits = bits
	h.Map = make([]int32, 1<<bits)
	for i := range h.Map {
		h.Map[i] = -1
	}
	h.Entries = make([]hashEntry, 0, size)
}

func (h *hashmap) Set(key uint32, val int) {
	if len(h.Entries) == 0x7fffffff {
		// max int32
		panic("too many entries")
	}
	low := key & (1<<h.Bits - 1)
	slot := &h.Map[low]
	if *slot < 0 {
		*slot = int32(len(h.Entries))
		h.Entries = append(h.Entries, hashEntry{-1, key, val})
		return
	}
	// chain entries in the same bucket.
	old := &h.Entries[*slot]
	if old.Key == key {
		if old.Val > val {
			old.Val = val
		}
		return
	}

	h.Entries = append(h.Entries, hashEntry{*slot, key, val})
	*slot = int32(len(h.Entries) - 1)
}

func (h *hashmap) Get(key uint32) (val int, ok bool) {
	low := key & (1<<h.Bits - 1)
	slot := h.Map[low]
	for slot >= 0 {
		e := &h.Entries[slot]
		if e.Key == key {
			return e.Val, true
		}
		slot = e.Next
	}
	return -1, false
}

