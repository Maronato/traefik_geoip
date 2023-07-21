// Taken from https://github.com/mmcloughlin/geohash/
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

package traefik_geoip //nolint:revive,stylecheck

// base32Encode bits of 64-bit word into a string.
func base32Encode(x uint64) string {
	alphabet := "0123456789bcdefghjkmnpqrstuvwxyz"
	b := [12]byte{}
	for i := 0; i < 12; i++ {
		b[11-i] = alphabet[x&0x1f]
		x >>= 5
	}
	return string(b[:])
}

// Encode the position of x within the range -r to +r as a 32-bit integer.
func encodeRange(x, r float64) uint32 {
	const exp232 = 4_294_967_296 // 2^32

	p := (x + r) / (2 * r)
	return uint32(p * exp232)
}

// Spread out the 32 bits of x into 64 bits, where the bits of x occupy even
// bit positions.
func spread(x uint32) uint64 {
	X := uint64(x)
	X = (X | (X << 16)) & 0x0000ffff0000ffff
	X = (X | (X << 8)) & 0x00ff00ff00ff00ff
	X = (X | (X << 4)) & 0x0f0f0f0f0f0f0f0f
	X = (X | (X << 2)) & 0x3333333333333333
	X = (X | (X << 1)) & 0x5555555555555555
	return X
}

// Interleave the bits of x and y. In the result, x and y occupy even and odd
// bitlevels, respectively.
func interleave(x, y uint32) uint64 {
	return spread(x) | (spread(y) << 1)
}

// encodeInt encodes the point (lat, lng) to a 64-bit integer geohash.
func encodeInt(lat, lng float64) uint64 {
	latInt := encodeRange(lat, 90)
	lngInt := encodeRange(lng, 180)
	return interleave(latInt, lngInt)
}

// EncodeIntWithPrecision encodes the point (lat, lng) to an integer with the
// specified number of bits.
func encodeIntWithPrecision(lat, lng float64, bits uint) uint64 {
	hash := encodeInt(lat, lng)
	return hash >> (64 - bits)
}

// encodeWithPrecision encodes the point (lat, lng) as a string geohash with
// the specified number of characters of precision (max 12).
func encodeWithPrecision(lat, lng float64, chars uint) string {
	bits := 5 * chars
	inthash := encodeIntWithPrecision(lat, lng, bits)
	enc := base32Encode(inthash)
	return enc[12-chars:]
}

// EncodeGeoHash the point (lat, lng) as a string geohash with the standard 12
// characters of precision.
func EncodeGeoHash(lat, lng float64) string {
	return encodeWithPrecision(lat, lng, 12)
}
