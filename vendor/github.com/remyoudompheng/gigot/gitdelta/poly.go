// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitdelta

// This file implements Rabin fingerprinting, a
// code defined by an irreducible polynomial mod 2.

// T and U are tables are seen in Git source code.
var _T, _U [256]uint32

// poly is an irreducible polynomial of degree 31, b_31 ... b_0
// representing b_31 X^31 + ... + b_1 X + b_0.
const poly = 0xab59b4d1

const degree = 31

// We use a hashing window of 16 bytes.
const _W = 16

func init() {
	initTables()
}

// initTables initializes tables T and U.
func initTables() {
	var bits [8]uint32 // bits[i] is X^(31+i) mod poly.
	p := uint32(poly &^ (1 << 31))
	for i := 0; i < 8; i++ {
		bits[i] = p
		if p>>31 == 1 {
			p ^= poly
		}
		p <<= 1
		if p>>31 == 1 {
			p ^= poly
		}
	}

	// Fill table T. T[i] = i * X^31 mod poly + X^31 if i is odd.
	for i := range _T {
		p := uint32(0)
		for j := 0; j < 8; j++ {
			if i&(1<<uint(j)) != 0 {
				p ^= bits[j]
			}
		}
		_T[i] = p | (uint32(i) << 31)
	}

	p = uint32(1)
	for i := 0; i < 8*_W; i++ {
		if i >= 8*_W-8 {
			// bits[i] = X^(8*Window-8+i) mod poly.
			bits[i-8*_W+8] = p
		}
		if p>>31 == 1 {
			p ^= poly
		}
		p <<= 1
		if p>>31 == 1 {
			p ^= poly
		}
	}

	// Fill table U. U[i] = i * X^(8*Window-8) mod poly.
	for i := range _U {
		p := uint32(0)
		for j := 0; j < 8; j++ {
			if i&(1<<uint(j)) != 0 {
				p ^= bits[j]
			}
		}
		_U[i] = p
	}
}

func hashRabin(s []byte) uint32 {
	var p uint32
	for _, b := range s {
		// p <- (p*X^8 + b) mod poly
		// t clears the upper bit and adds the degree >= 31
		// part of p*X^8.
		t := _T[byte(p>>(degree-8))]
		p = (p << 8) ^ uint32(b) ^ t
	}
	return p
}
