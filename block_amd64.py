import peachpy.x86_64

# This 8-way SHA256 is based on Daniel J. Bernstein's crypto_hash_sha256
# and the 8-way SHA256 implementation in reference code for SPHINCS+.


# Applies VPSHUFB to the multiple-of-eight uint32_t slice
def vpshufb(offset, todo, mask):
    offset = GeneralPurposeRegister64()

    with Loop() as loop:
        tmp = YMMRegister()
        VMOVDQU(tmp, [offset])
        VPSHUFB(tmp, tmp, [mask])
        VMOVDQU([offset], tmp)
        ADD(offset, 32)
        SUB(todo, 8)
        JNZ(loop.begin)

# Applies VPSHUFB to the multiple-of-eight uint32_t slice
buf = Argument(ptr(uint32_t))
buf_len = Argument(size_t)
mask = Argument(ptr(uint8_t))
with Function("vpshufb", (buf, buf_len, mask),
            target=uarch.haswell) as function:
    offset = GeneralPurposeRegister64()
    todo = GeneralPurposeRegister64()
    reg_mask= GeneralPurposeRegister64()

    LOAD.ARGUMENT(todo, buf_len)
    LOAD.ARGUMENT(offset, buf)
    LOAD.ARGUMENT(reg_mask, mask)

    vpshufb(offset, todo, reg_mask)

    RETURN ()

# Transposes the 8x8 table of *uint32 in s
def transpose(s):
    tmp0 = [YMMRegister() for i in xrange(8)]
    tmp1 = [YMMRegister() for i in xrange(8)]
    for i in xrange(4):
        VMOVDQU(tmp0[i*2], s[i*2])
        VPUNPCKLDQ(tmp0[i*2], tmp0[i*2], s[i*2+1])
        VMOVDQU(tmp0[i*2+1], s[i*2])
        VPUNPCKHDQ(tmp0[i*2+1], tmp0[i*2+1], s[i*2+1])
    VPUNPCKLQDQ(tmp1[0], tmp0[0], tmp0[2])
    VPUNPCKHQDQ(tmp1[1], tmp0[0], tmp0[2])
    VPUNPCKLQDQ(tmp1[2], tmp0[1], tmp0[3])
    VPUNPCKHQDQ(tmp1[3], tmp0[1], tmp0[3])
    VPUNPCKLQDQ(tmp1[4], tmp0[4], tmp0[6])
    VPUNPCKHQDQ(tmp1[5], tmp0[4], tmp0[6])
    VPUNPCKLQDQ(tmp1[6], tmp0[5], tmp0[7])
    VPUNPCKHQDQ(tmp1[7], tmp0[5], tmp0[7])
    VPERM2I128(tmp0[0], tmp1[0], tmp1[4], 0x20)
    VPERM2I128(tmp0[1], tmp1[1], tmp1[5], 0x20)
    VPERM2I128(tmp0[2], tmp1[2], tmp1[6], 0x20)
    VPERM2I128(tmp0[3], tmp1[3], tmp1[7], 0x20)
    VPERM2I128(tmp0[4], tmp1[0], tmp1[4], 0x31)
    VPERM2I128(tmp0[5], tmp1[1], tmp1[5], 0x31)
    VPERM2I128(tmp0[6], tmp1[2], tmp1[6], 0x31)
    VPERM2I128(tmp0[7], tmp1[3], tmp1[7], 0x31)
    for i in xrange(8):
        VMOVDQU(s[i], tmp0[i])


# Transposes the 8x8 table of uint32s
table = Argument(ptr(uint32_t))
with Function("transpose", (table,), target=uarch.haswell) as function:
    reg_table = GeneralPurposeRegister64()
    LOAD.ARGUMENT(reg_table, table)
    s = [[reg_table + i*32] for i in xrange(8)]
    transpose(s)
    RETURN()

# helpers
def avx2_rotr32(res, x, y):
    tmp = YMMRegister()
    VPSRLD(res, x, y)
    VPSLLD(tmp, x, 32-y)
    VPOR(res, tmp, res)

s = Argument(ptr(uint32_t))
data = Argument(ptr(ptr(uint8_t)))
nblocks = Argument(size_t)
K = Argument(ptr(uint32_t))
mask = Argument(ptr(uint8_t))
with Function("block", (s, data, nblocks, K, mask), target=uarch.haswell) as function:
    reg_s = GeneralPurposeRegister64()
    reg_data = GeneralPurposeRegister64()
    reg_K = GeneralPurposeRegister64()
    reg_mask = GeneralPurposeRegister64()
    reg_nblocks = GeneralPurposeRegister64()

    LOAD.ARGUMENT(reg_s, s)
    LOAD.ARGUMENT(reg_data, data)
    LOAD.ARGUMENT(reg_nblocks, nblocks)
    LOAD.ARGUMENT(reg_K, K)
    LOAD.ARGUMENT(reg_mask, mask)

    reg_sp_save = GeneralPurposeRegister64()
    reg_data_offset = GeneralPurposeRegister64()
    MOV(reg_data_offset, 0)

    # Align stack to 32byte boundary
    MOV(reg_sp_save, registers.rsp)
    AND(registers.rsp, 0xffffffffffffffe0)
    SUB(registers.rsp, 0x20)

    SUB(registers.rsp, 512)
    w = [[registers.rsp + 32*i] for i in xrange(16)]

    with Loop() as loop:
        tmp = YMMRegister()

        # Load data into w
        for i in xrange(8):
            reg_ddata = GeneralPurposeRegister64()
            MOV(reg_ddata, [reg_data + 8*i])
            ADD(reg_ddata, reg_data_offset)
            VMOVDQU(tmp, [reg_ddata])
            VPSHUFB(tmp, tmp, [reg_mask])
            VMOVDQU(w[i], tmp)
            VMOVDQU(tmp, [reg_ddata + 32])
            VPSHUFB(tmp, tmp, [reg_mask])
            VMOVDQU(w[i+8], tmp)

        transpose(w)
        transpose(w[8:])

        # Copy state into ss
        ss = [YMMRegister() for i in xrange(8)]
        a, b, c, d, e, f, g, h = ss

        T1, T2 = YMMRegister(), YMMRegister()

        for i in xrange(8):
            VMOVDQU(ss[i], [reg_s+32*i])

        sig0, sig1 = YMMRegister(), YMMRegister()
        ch, maj = sig0, sig1

        for i in xrange(64):
        # for i in xrange(1):
            if i != 0 and i % 16 == 0:
                # expand
                for j in xrange(16):
                    w0 = w[j]
                    w14 = w[(j+14)%16]
                    w9 = w[(j+9)%16]
                    w1 = w[(j+1)%16]

                    reg_w0 = YMMRegister()
                    VMOVDQU(reg_w0, w0)

                    # w0 = sigma1(w14) + w9 + sigma0(w1) + w0
                    #  sub1: sig1 = sigma1(w14) = ROTR(w14,17) ^ ROTR(w14,19) ^ SHR(w14,10)
                    reg_w14 = YMMRegister()
                    VMOVDQU(reg_w14, w14)
                    avx2_rotr32(sig1, reg_w14, 17)
                    avx2_rotr32(tmp, reg_w14, 19)
                    VPXOR(sig1, sig1, tmp)
                    VPSRLD(tmp, reg_w14, 10)
                    VPXOR(sig1, sig1, tmp)

                    #  sub2: sig = sig + sigma0(w1) = ROTR(w1, 7) ^ ROTR(w1,18) ^ SHR(w1, 3)
                    reg_w1 = YMMRegister()
                    VMOVDQU(reg_w1, w1)
                    avx2_rotr32(sig0, reg_w1, 7)
                    avx2_rotr32(tmp, reg_w1, 18)
                    VPXOR(sig0, sig0, tmp)
                    VPSRLD(tmp, reg_w1, 3)
                    VPXOR(sig0, sig0, tmp)

                    VPADDD(reg_w0, reg_w0, sig1)
                    VPADDD(reg_w0, reg_w0, sig0)
                    VPADDD(reg_w0, reg_w0, w9)
                    VMOVDQU(w0, reg_w0)


            # Subround (F)
            # Part 1: T1 = h + Sigma1(e) + Ch(e,f,g) + k + w;
            #  sub1: T1 = Sigma1(e) = ROTR(e, 6) ^ ROTR(e,11) ^ ROTR(e,25)
            avx2_rotr32(T1, e, 6)
            avx2_rotr32(tmp, e, 11)
            VPXOR(T1, T1, tmp)
            avx2_rotr32(tmp, e, 25)
            VPXOR(T1, T1, tmp)

            # sub2: Ch = Ch(e,f,g) = (e & f) ^ (~e & g)
            VPAND(ch, e, f)
            VPCMPEQD(tmp, tmp, tmp) # tmp = 0xfff..f
            VPXOR(tmp, tmp, e) # tmp = ~e
            VPAND(tmp, tmp, g) # tmp = ~e & g
            VPXOR(ch, ch, tmp)

            # sub3: T1 = T1 + ch + h + k + w
            VPADDD(T1, T1, ch)
            VPADDD(T1, T1, h)
            VPBROADCASTD(tmp, [reg_K + 4*i])
            VPADDD(T1, T1, tmp)
            VMOVDQU(tmp, w[i % 16])
            VPADDD(T1, T1, tmp)

            # Part 2: T2 = Sigma0(a) + Maj(a,b,c)
            #  sub1: T2 = Sigma0(a) = ROTR(a, 2) ^ ROTR(a,13) ^ ROTR(a,22)
            avx2_rotr32(T2, a, 2)
            avx2_rotr32(tmp, a, 13)
            VPXOR(T2, T2, tmp)
            avx2_rotr32(tmp, a, 22)
            VPXOR(T2, T2, tmp)

            # sub2: maj = Maj(a,b,c) = (a&b) ^ (a&c) ^ (b&c)
            VPAND(maj, a, b)
            VPAND(tmp, a, c)
            VPXOR(maj, maj, tmp)
            VPAND(tmp, b, c)
            VPXOR(maj, maj, tmp)

            VPADDD(T2, T2, maj)
            
            # Shuffling
            VMOVDQU(h, g)
            VMOVDQU(g, f)
            VMOVDQU(f, e)
            VPADDD(e, d, T1)
            VMOVDQU(d, c)
            VMOVDQU(c, b)
            VMOVDQU(b, a)
            VPADDD(a, T1, T2)

        # Write out state
        for i in xrange(8):
            VPADDD(ss[i], ss[i], [reg_s+32*i])
            VMOVDQU([reg_s+32*i], ss[i])

        ADD(reg_data_offset, 64)
        SUB(reg_nblocks, 1)
        JNZ(loop.begin)

    # Reset stack
    MOV(registers.rsp, reg_sp_save)

    RETURN ()
