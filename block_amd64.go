//go:generate python -m peachpy.x86_64 block_amd64.py -S -o block_amd64.s -mabi=goasm

package sha256x8

import (
	"github.com/intel-go/cpuid"
)

var available = cpuid.EnabledAVX && cpuid.HasExtendedFeature(cpuid.AVX2)

// Computes one block of SHA256
func block(s *uint32, data **byte, nblocks int, K *uint32, bsMask *byte)

// Transposes the given 8x8 table of uint32s
func transpose(table *uint32)

// Applies VPSHUFB to a multiple-of-eight uint32 slice
func vpshufb(buf *uint32, buf_len int, mask *byte)

// Swaps a slice of uint32s between little and big endian
func byteswap(table []uint32) {
	vpshufb(&table[0], len(table), &byteswapMask[0])
}
