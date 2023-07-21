// Take from https://github.com/mmcloughlin/geohash/
// The MIT License (MIT)

// Copyright (c) 2015 Michael McLoughlin

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package traefik_geoip

// invalid is a placeholder for invalid character decodings.
const invalid = 0xff

// encoding encapsulates an encoding defined by a given base32 alphabet.
type encoding struct {
	encode string
	decode [256]byte
}

// newEncoding constructs a new encoding defined by the given alphabet,
// which must be a 32-byte string.
func newEncoding(encoder string) *encoding {
	e := new(encoding)
	e.encode = encoder
	for i := 0; i < len(e.decode); i++ {
		e.decode[i] = invalid
	}
	for i := 0; i < len(encoder); i++ {
		e.decode[encoder[i]] = byte(i)
	}
	return e
}

// ValidByte reports whether b is part of the encoding.
func (e *encoding) ValidByte(b byte) bool {
	return e.decode[b] != invalid
}

// Decode string into bits of a 64-bit word. The string s may be at most 12
// characters.
func (e *encoding) Decode(s string) uint64 {
	x := uint64(0)
	for i := 0; i < len(s); i++ {
		x = (x << 5) | uint64(e.decode[s[i]])
	}
	return x
}

// Encode bits of 64-bit word into a string.
func (e *encoding) Encode(x uint64) string {
	b := [12]byte{}
	for i := 0; i < 12; i++ {
		b[11-i] = e.encode[x&0x1f]
		x >>= 5
	}
	return string(b[:])
}

// Base32Encoding with the Geohash alphabet.
var base32encoding = newEncoding("0123456789bcdefghjkmnpqrstuvwxyz")
