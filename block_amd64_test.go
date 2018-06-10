package sha256x8

import (
	"math/rand"
	"testing"
)

func TestByteSwapAvx2(t *testing.T) {
	var t1 [8 * 10]uint32
	var t2 [8 * 10]uint32
	for i := 0; i < len(t1); i++ {
		t1[i] = rand.Uint32()
	}
	copy(t2[:], t1[:])
	byteswap(t1[:])
	for i := 0; i < len(t1); i++ {
		if t1[i] != (((t2[i] >> 24) & 255) |
			((t2[i] >> 8) & (255 << 8)) |
			((t2[i] << 8) & (255 << 16)) |
			((t2[i] << 24) & (255 << 24))) {
			t.Fatal()
		}
	}
}

func TestBlockAvx2(t *testing.T) {
	var s, s2 [64]uint32
	var data, data2 [64 * 8]byte
	rand.Read(data[:])

	for i := 0; i < 64; i++ {
		s[i] = rand.Uint32()
		s2[i] = s[i]
	}

	copy(data2[:], data[:])

	pData := []*byte{
		&data2[0],
		&data2[64],
		&data2[128],
		&data2[192],
		&data2[256],
		&data2[320],
		&data2[384],
		&data2[448],
	}

	blockGeneric(s2[:], data[:])
	transpose(&s[0])
	block(&s[0], &pData[0], 1, &_K[0], &byteswapMask[0])
	transpose(&s[0])
	for i := 0; i < 64; i++ {
		if s[i] != s2[i] {
			t.Logf("avx: %X", s)
			t.Logf("gen: %X", s2)
			t.Fatal()
		}
	}
}

func TestTransposeAvx2(t *testing.T) {
	var table [64]uint32
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			k := i*8 + j
			v := uint32(k + (k << 8) + (k << 16) + (k << 24))
			table[k] = v
		}
	}
	transpose(&table[0])
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			k := i*8 + j
			v := uint32(k + (k << 8) + (k << 16) + (k << 24))
			if table[(j*8)+i] != v {
				t.Fatalf("%v != %v", table[(j*8)+i], v)
			}
		}
	}
}

func BenchmarkBlockAvx2(b *testing.B) {
	var s [64]uint32
	var d [512]byte
	pD := []*byte{
		&d[0],
		&d[64],
		&d[128],
		&d[192],
		&d[256],
		&d[320],
		&d[384],
		&d[448],
	}
	for n := 0; n < b.N; n++ {
		block(&s[0], &pD[0], 1, &_K[0], &byteswapMask[0])
	}
}

func BenchmarkByteSwap(b *testing.B) {
	var t [16]uint32
	for n := 0; n < b.N; n++ {
		byteswap(t[:])
	}
}

func BenchmarkTranspose(b *testing.B) {
	var t [8]uint32
	for n := 0; n < b.N; n++ {
		transpose(&t[0])
	}
}

func blockGeneric(h []uint32, p []byte) {
	for i := 0; i < 8; i++ {
		blockGenericX1(h[i*8:(i+1)*8], p[i*64:(i+1)*64])
	}
}

func blockGenericX1(hs []uint32, p []byte) {
	var w [64]uint32
	h0, h1, h2, h3, h4, h5, h6, h7 := hs[0], hs[1], hs[2], hs[3], hs[4], hs[5], hs[6], hs[7]
	// Can interlace the computation of w with the
	// rounds below if needed for speed.
	for i := 0; i < 16; i++ {
		j := i * 4
		w[i] = uint32(p[j])<<24 | uint32(p[j+1])<<16 | uint32(p[j+2])<<8 | uint32(p[j+3])
	}
	for i := 16; i < 64; i++ {
		v1 := w[i-2]
		t1 := (v1>>17 | v1<<(32-17)) ^ (v1>>19 | v1<<(32-19)) ^ (v1 >> 10)
		v2 := w[i-15]
		t2 := (v2>>7 | v2<<(32-7)) ^ (v2>>18 | v2<<(32-18)) ^ (v2 >> 3)
		w[i] = t1 + w[i-7] + t2 + w[i-16]
	}

	a, b, c, d, e, f, g, h := h0, h1, h2, h3, h4, h5, h6, h7

	for i := 0; i < 64; i++ {
		// for i := 0; i < 1; i++ {
		t1 := h + ((e>>6 | e<<(32-6)) ^ (e>>11 | e<<(32-11)) ^ (e>>25 | e<<(32-25))) + ((e & f) ^ (^e & g)) + _K[i] + w[i]

		t2 := ((a>>2 | a<<(32-2)) ^ (a>>13 | a<<(32-13)) ^ (a>>22 | a<<(32-22))) + ((a & b) ^ (a & c) ^ (b & c))

		h = g
		g = f
		f = e
		e = d + t1
		d = c
		c = b
		b = a
		a = t1 + t2
	}

	h0 += a
	h1 += b
	h2 += c
	h3 += d
	h4 += e
	h5 += f
	h6 += g
	h7 += h

	hs[0], hs[1], hs[2], hs[3], hs[4], hs[5], hs[6], hs[7] = h0, h1, h2, h3, h4, h5, h6, h7
}
