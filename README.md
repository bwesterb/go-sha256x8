sha256x8
========

This pure Go package implements an eight-way SHA256 using AVX2 instructions.

Basic usage
-----------

The following

```go
data := [8][]byte{
    []byte("Eight pieces of data"),
    []byte("which all must have "),
    []byte("the same length     "),
    []byte("                    "),
    []byte("Eight pieces of data"),
    []byte("which all must have "),
    []byte("the same length     "),
    []byte("                    "),
}
out := [8][32]byte{}
sha256x8.SumsInto(data,
    [8][]byte{
        out[0][:], out[1][:], out[2][:], out[3][:],
        out[4][:], out[5][:], out[6][:], out[7][:],
    })
fmt.Printf("%x\n", out)
```

will output
```
[38ce9135c02f5ee3cfc527e9d33d6cc7c5f387ee0f272af21d73b14c79c8bad4 c0aa3c27409dcebf899fb5eac8fd5ea098c22b3e2400b79f1a1128df83954aa9 928c112e5a5d9eac1d6100d603942ec4a8b1002b04449cc05fb3388bb4dde9b6 1e8a105dbcab2f5d6b30a670c0ff91942f4db62401e669331037101e94198250 38ce9135c02f5ee3cfc527e9d33d6cc7c5f387ee0f272af21d73b14c79c8bad4 c0aa3c27409dcebf899fb5eac8fd5ea098c22b3e2400b79f1a1128df83954aa9 928c112e5a5d9eac1d6100d603942ec4a8b1002b04449cc05fb3388bb4dde9b6 1e8a105dbcab2f5d6b30a670c0ff91942f4db62401e669331037101e94198250]
```

[See godoc for more documentation](https://godoc.org/github.com/bwesterb/go-sha256x8).

Acknowledgement
---------------
This package is based on the [SPHINCS+ AVX2 sha256x8 code](
https://github.com/sphincs/sphincsplus) and Bernstein's `crypto_hash_sha256`.

Contributing
------------

There is still quite a lot to do:

 - Detect presence of AVX2 and use a fallback if it is not available.
 - SHA256NI backend
 - AVX512 backend
 - AVX backend
 - ARM SIMD backend
