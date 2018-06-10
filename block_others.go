// +build !amd64

package sha256x8

var available = false

func block(s *uint32, data *uint32, K *uint32, bsMask *byte) {
	panic("GOARCH not supported")
}

// Transposes the given 8x8 table of uint32s
func transpose(table *uint32) {
	panic("GOARCH not supported")
}

// Applies VPSHUFB to a multiple-of-eight uint32 slice
func vpshufb(buf *uint32, buf_len int, mask *byte) {
	panic("GOARCH not supported")
}

// Swaps a slice of uint32s between little and big endian
func byteswap(table []uint32) {
	panic("GOARCH not supported")
}
