// Pure Go implementation of eightway sha256 using AVX2 instructions.
//
// Check sha256x8.Available() before calling any other function.
package sha256x8

import (
	"reflect"
	"unsafe"
)

// Returns whether the current architecture is supported.
//
// If false is returned, calling any of the other functions might result in
// panics or illegal instruction exceptions.
func Available() bool {
	return available
}

// Represents the state of eight sha256 hashes.
//
// Either call New() to create an instance, or call Reset() on a zero
// instance on the stack before use.
type Digest struct {
	// the checksum state: a0 b0 c0 ... h0 a1 b1 c1 .. h1 ... h7
	s [64]uint32

	// total number of bytes consumed
	l uint64

	// partially filled transposed block
	p [512]byte

	// number of bytes of data in p per hash (ie. between 0 and 64)
	pn int
}

var initialState = [64]uint32{
	0x6A09E667, 0x6A09E667, 0x6A09E667, 0x6A09E667,
	0x6A09E667, 0x6A09E667, 0x6A09E667, 0x6A09E667,
	0xBB67AE85, 0xBB67AE85, 0xBB67AE85, 0xBB67AE85,
	0xBB67AE85, 0xBB67AE85, 0xBB67AE85, 0xBB67AE85,
	0x3C6EF372, 0x3C6EF372, 0x3C6EF372, 0x3C6EF372,
	0x3C6EF372, 0x3C6EF372, 0x3C6EF372, 0x3C6EF372,
	0xA54FF53A, 0xA54FF53A, 0xA54FF53A, 0xA54FF53A,
	0xA54FF53A, 0xA54FF53A, 0xA54FF53A, 0xA54FF53A,
	0x510E527F, 0x510E527F, 0x510E527F, 0x510E527F,
	0x510E527F, 0x510E527F, 0x510E527F, 0x510E527F,
	0x9B05688C, 0x9B05688C, 0x9B05688C, 0x9B05688C,
	0x9B05688C, 0x9B05688C, 0x9B05688C, 0x9B05688C,
	0x1F83D9AB, 0x1F83D9AB, 0x1F83D9AB, 0x1F83D9AB,
	0x1F83D9AB, 0x1F83D9AB, 0x1F83D9AB, 0x1F83D9AB,
	0x5BE0CD19, 0x5BE0CD19, 0x5BE0CD19, 0x5BE0CD19,
	0x5BE0CD19, 0x5BE0CD19, 0x5BE0CD19, 0x5BE0CD19}

// Resets the state.
func (d *Digest) Reset() {
	copy(d.s[:], initialState[:])
	d.l = 0
	d.pn = 0
}

// Create 8 new SHA256 digests.
func New() *Digest {
	var d Digest
	d.Reset()
	return &d
}

func castByteSliceToUint32Slice(buf []byte) []uint32 {
	header := *(*reflect.SliceHeader)(unsafe.Pointer(&buf))
	header.Len /= 4
	header.Cap /= 4
	return *(*[]uint32)(unsafe.Pointer(&header))
}

// Feed data[i] to the i'th sha256.
//
// Requires each data[i] to be of the same length.
func (d *Digest) Write(data [8][]byte) {
	d.l += uint64(len(data[0]))
	for len(data[0]) > 0 {
		if len(data[0]) >= 64 && d.pn == 0 {
			nblocks := len(data[0]) / 64
			pPtr := [8]*byte{
				&data[0][0], &data[1][0], &data[2][0], &data[3][0],
				&data[4][0], &data[5][0], &data[6][0], &data[7][0],
			}
			block(&d.s[0], &pPtr[0], nblocks, &_K[0], &byteswapMask[0])
			for i := 0; i < 8; i++ {
				data[i] = data[i][64*nblocks:]
			}
			continue
		}
		pn2 := len(data[0]) + d.pn
		toCopy := len(data[0])
		if pn2 >= 64 {
			toCopy = len(data[0]) - (pn2 - 64)
			pn2 = 0
		}
		for i := 0; i < 8; i++ {
			copy(d.p[i*64+d.pn:], data[i][:toCopy])
			data[i] = data[i][toCopy:]
		}
		d.pn = pn2
		if pn2 == 0 {
			pPtr := [8]*byte{
				&d.p[0], &d.p[64], &d.p[128], &d.p[192],
				&d.p[256], &d.p[320], &d.p[384], &d.p[448],
			}
			block(&d.s[0], &pPtr[0], 1, &_K[0], &byteswapMask[0])
		}
	}
}

// Writes the i'th sha256 into out[i].  Invalidates d.
func (d *Digest) SumsInto(out [8][]byte) {
	var tmp [64]byte
	lb := d.l << 3
	if d.pn < 56 {
		// Padding fits in current block
		tmp[d.pn] = 0x80
		for i := uint(0); i < 8; i++ {
			tmp[64-i-1] = byte(lb >> (8 * i))
		}
		for i := 0; i < 8; i++ {
			copy(d.p[64*i+d.pn:64*(i+1)], tmp[d.pn:64])
		}
		pPtr := [8]*byte{
			&d.p[0], &d.p[64], &d.p[128], &d.p[192],
			&d.p[256], &d.p[320], &d.p[384], &d.p[448],
		}
		block(&d.s[0], &pPtr[0], 1, &_K[0], &byteswapMask[0])
	} else {
		var padding []byte
		tmp[0] = 0x80
		padding = tmp[:64+56-d.pn]
		d.Write([8][]byte{padding, padding, padding, padding, padding, padding, padding, padding})
		for i := uint(0); i < 8; i++ {
			tmp[i] = byte(lb >> (56 - 8*i))
		}
		padding = tmp[:8]
		d.Write([8][]byte{padding, padding, padding, padding, padding, padding, padding, padding})
		if d.pn != 0 {
			panic("d.pn != 0")
		}
	}
	byteswap(d.s[:])
	transpose(&d.s[0])
	for i := 0; i < 8; i++ {
		copy(castByteSliceToUint32Slice(out[i]), d.s[i*8:(i+1)*8])
	}
}

// Returns the sha256s. Invalidates d.
func (d *Digest) Sums() [8][]byte {
	var buf [256]byte
	var ret [8][]byte
	for i := 0; i < 8; i++ {
		ret[i] = buf[i*32 : (i+1)*32]
	}
	d.SumsInto(ret)
	return ret
}

// Computes the sha256 sums of the buffers, which must be of the same length,
// and writes these to the provided byteslices.
func SumsInto(data [8][]byte, out [8][]byte) {
	var d Digest
	d.Reset()
	d.Write(data)
	d.SumsInto(out)
}
