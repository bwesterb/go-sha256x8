package sha256x8_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"math/rand"
	"testing"

	"github.com/bwesterb/go-sha256x8"
)

func TestEmptyDigest(t *testing.T) {
	var d sha256x8.Digest
	d.Reset()
	sums := d.Sums()
	for i := 0; i < 8; i++ {
		if hex.EncodeToString(sums[i]) != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
			t.Fatal()
		}
	}
}

func TestSingleByte(t *testing.T) {
	var d sha256x8.Digest
	d.Reset()
	d.Write([8][]byte{[]byte{0}, []byte{1}, []byte{2}, []byte{3}, []byte{4},
		[]byte{5}, []byte{6}, []byte{7}})
	sums := d.Sums()
	expected := []string{
		"6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
		"4bf5122f344554c53bde2ebb8cd2b7e3d1600ad631c385a5d7cce23c7785459a",
		"dbc1b4c900ffe48d575b5da5c638040125f65db0fe3e24494b76ea986457d986",
		"084fed08b978af4d7d196a7446a86b58009e636b611db16211b65a9aadff29c5",
		"e52d9c508c502347344d8c07ad91cbd6068afc75ff6292f062a09ca381c89e71",
		"e77b9a9ae9e30b0dbdb6f510a264ef9de781501d7b6b92ae89eb059c5ab743db",
		"67586e98fad27da0b9968bc039a1ef34c939b9b8e523a8bef89d478608c5ecf6",
		"ca358758f6d27e6cf45272937977a748fd88391db679ceda7dc7bf1f005ee879",
	}
	for i := 0; i < 8; i++ {
		if expected[i] != hex.EncodeToString(sums[i]) {
			t.Fatalf("%d: %x != %s", i, sums[i], expected[i])
		}
	}
}

func TestFuzz(t *testing.T) {
	var refs [8]hash.Hash
	var d sha256x8.Digest
	for n := 0; n < 1000; n++ {
		d.Reset()
		for i := 0; i < 8; i++ {
			refs[i] = sha256.New()
		}
		for bit := 0; bit < 100; bit++ {
			l := rand.Int31n(128)
			var bufs [8][]byte
			for i := 0; i < 8; i++ {
				bufs[i] = make([]byte, l)
				rand.Read(bufs[i])
				refs[i].Write(bufs[i])
			}
			d.Write(bufs)
		}
		sums := d.Sums()
		for i := 0; i < 8; i++ {
			if !bytes.Equal(refs[i].Sum([]byte{}), sums[i]) {
				t.Fatal()
			}
		}
	}
}

func BenchmarkSha256x8On1B(b *testing.B) {
	benchmarkSha256x8(1, b)
}

func BenchmarkSha256x8On32B(b *testing.B) {
	benchmarkSha256x8(32, b)
}

func BenchmarkSha256x8On64B(b *testing.B) {
	benchmarkSha256x8(64, b)
}

func BenchmarkSha256x8On96B(b *testing.B) {
	benchmarkSha256x8(96, b)
}

func BenchmarkSha256x8On256B(b *testing.B) {
	benchmarkSha256x8(256, b)
}

func BenchmarkSha256x8On10240B(b *testing.B) {
	benchmarkSha256x8(10240, b)
}

func BenchmarkSha256On1B(b *testing.B) {
	benchmarkSha256(1, b)
}

func BenchmarkSha256On32B(b *testing.B) {
	benchmarkSha256(32, b)
}

func BenchmarkSha256On64B(b *testing.B) {
	benchmarkSha256(64, b)
}

func BenchmarkSha256On96B(b *testing.B) {
	benchmarkSha256(96, b)
}

func BenchmarkSha256On256B(b *testing.B) {
	benchmarkSha256(256, b)
}

func BenchmarkSha256On10240B(b *testing.B) {
	benchmarkSha256(10240, b)
}

func benchmarkSha256x8(dataLen int, b *testing.B) {
	buf := make([]byte, dataLen)
	var d sha256x8.Digest
	var out [32]byte
	bufs := [8][]byte{buf, buf, buf, buf, buf, buf, buf, buf}
	outs := [8][]byte{out[:], out[:], out[:], out[:], out[:], out[:], out[:], out[:]}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		d.Reset()
		d.Write(bufs)
		d.SumsInto(outs)
	}
}

func benchmarkSha256(dataLen int, b *testing.B) {
	d := make([]byte, dataLen)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sha256.Sum256(d)
		sha256.Sum256(d)
		sha256.Sum256(d)
		sha256.Sum256(d)
		sha256.Sum256(d)
		sha256.Sum256(d)
		sha256.Sum256(d)
		sha256.Sum256(d)
	}
}
