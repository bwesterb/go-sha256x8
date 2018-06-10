// +build amd64
//go:generate python -m peachpy.x86_64 block_amd64.py -S -o block_amd64.s -mabi=goasm

package sha256x8

var byteswapMask = [32]uint8{
	0x3, 0x2, 0x1, 0x0, 0x7, 0x6, 0x5, 0x4, 0xb, 0xa, 0x9, 0x8, 0xf, 0xe, 0xd, 0xc,
	0x3, 0x2, 0x1, 0x0, 0x7, 0x6, 0x5, 0x4, 0xb, 0xa, 0x9, 0x8, 0xf, 0xe, 0xd, 0xc,
}

// Computes one block of SHA256
func block(s *uint32, data *uint32, K *uint32, bsMask *byte)

// Transposes the given 8x8 table of uint32s
func transpose(table *uint32)

// Applies VPSHUFB to a multiple-of-eight uint32 slice
func vpshufb(buf *uint32, buf_len int, mask *byte)

// Swaps a slice of uint32s between little and big endian
func byteswap(table []uint32) {
	vpshufb(&table[0], len(table), &byteswapMask[0])
}
