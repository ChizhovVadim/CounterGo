//go:build arm64

#include "textflag.h"

// ARM64 NEON for float32: 4 elements per 128-bit register
// All vector instructions use WORD opcodes since Go's ARM64 assembler
// doesn't support NEON mnemonics directly.

// Opcode formulas for float32 (4S arrangement):
// FADD Vd.4S, Vn.4S, Vm.4S: 0x4E20D400 | (Vm << 16) | (Vn << 5) | Vd
// FSUB Vd.4S, Vn.4S, Vm.4S: 0x4EA0D400 | (Vm << 16) | (Vn << 5) | Vd
// FMUL Vd.4S, Vn.4S, Vm.4S: 0x6E20DC00 | (Vm << 16) | (Vn << 5) | Vd
// FDIV Vd.4S, Vn.4S, Vm.4S: 0x6E20FC00 | (Vm << 16) | (Vn << 5) | Vd
// FMIN Vd.4S, Vn.4S, Vm.4S: 0x4EA0F400 | (Vm << 16) | (Vn << 5) | Vd
// FMAX Vd.4S, Vn.4S, Vm.4S: 0x4E20F400 | (Vm << 16) | (Vn << 5) | Vd
// FABS Vd.4S, Vn.4S:        0x4EA0F800 | (Vn << 5) | Vd
// FNEG Vd.4S, Vn.4S:        0x6EA0F800 | (Vn << 5) | Vd
// FMLA Vd.4S, Vn.4S, Vm.4S: 0x4E20CC00 | (Vm << 16) | (Vn << 5) | Vd
// FADDP Vd.4S, Vn.4S, Vm.4S: 0x6E20D400 | (Vm << 16) | (Vn << 5) | Vd
// FADDP Sd, Vn.2S:          0x7E30D800 | (Vn << 5) | Vd

// func dotProductNEON(a, b []float32) float32
// Handles mismatched slice lengths: uses min(len(a), len(b)).
TEXT ·dotProductNEON(SB), NOSPLIT, $0-52
    MOVD a_base+0(FP), R0
    MOVD a_len+8(FP), R2
    MOVD b_len+32(FP), R3
    CMP R3, R2
    CSEL LT, R2, R3, R2        // R2 = min(len(a), len(b))
    MOVD b_base+24(FP), R1

    VEOR V0.B16, V0.B16, V0.B16
    VEOR V1.B16, V1.B16, V1.B16

    // Process 8 elements (2 NEON ops) per iteration
    LSR $3, R2, R3
    CBZ R3, dot32_remainder4

dot32_loop8:
    VLD1.P 16(R0), [V2.S4]
    VLD1.P 16(R0), [V3.S4]
    VLD1.P 16(R1), [V4.S4]
    VLD1.P 16(R1), [V5.S4]
    WORD $0x4E24CC40           // FMLA V0.4S, V2.4S, V4.4S
    WORD $0x4E25CC61           // FMLA V1.4S, V3.4S, V5.4S
    SUB $1, R3
    CBNZ R3, dot32_loop8

    // Combine accumulators: V0 = V0 + V1
    WORD $0x4E21D400           // FADD V0.4S, V0.4S, V1.4S

dot32_remainder4:
    AND $7, R2, R3
    LSR $2, R3, R4
    CBZ R4, dot32_remainder

    VLD1.P 16(R0), [V2.S4]
    VLD1.P 16(R1), [V4.S4]
    WORD $0x4E24CC40           // FMLA V0.4S, V2.4S, V4.4S

dot32_remainder:
    AND $3, R3, R4
    CBZ R4, dot32_reduce

    // Must reduce vector FIRST before scalar ops (scalar ops zero upper V bits)
    WORD $0x6E20D400           // FADDP V0.4S, V0.4S, V0.4S
    WORD $0x7E30D800           // FADDP S0, V0.2S

dot32_scalar:
    FMOVS (R0), F2
    FMOVS (R1), F4
    FMADDS F4, F0, F2, F0      // F0 = F2 * F4 + F0 (Go syntax: Fm, Fa, Fn, Fd)
    ADD $4, R0
    ADD $4, R1
    SUB $1, R4
    CBNZ R4, dot32_scalar

    FMOVS F0, ret+48(FP)
    RET

dot32_reduce:
    // Horizontal sum of V0.4S -> S0 when no scalar remainder
    WORD $0x6E20D400           // FADDP V0.4S, V0.4S, V0.4S
    WORD $0x7E30D800           // FADDP S0, V0.2S

    FMOVS F0, ret+48(FP)
    RET


// func reluNEON(dst, src []float32)
// Computes ReLU: dst[i] = max(0, src[i])
TEXT ·reluNEON(SB), NOSPLIT, $0-48
    MOVD dst_base+0(FP), R0
    MOVD dst_len+8(FP), R3
    MOVD src_base+24(FP), R1

    // Create zero vector
    VEOR V30.B16, V30.B16, V30.B16    // V30 = {0, 0, 0, 0}

    // Process 4 elements per iteration
    LSR $2, R3, R4
    CBZ R4, relu32_neon_scalar

relu32_neon_loop4:
    VLD1.P 16(R1), [V0.S4]            // V0 = x
    WORD $0x4E3EF401                  // FMAX V1.4S, V0.4S, V30.4S -> V1 = max(x, 0)
    VST1.P [V1.S4], 16(R0)            // store result

    SUB $1, R4
    CBNZ R4, relu32_neon_loop4

relu32_neon_scalar:
    AND $3, R3
    CBZ R3, relu32_neon_done

relu32_neon_scalar_loop:
    FMOVS (R1), F0                    // F0 = x
    FMOVS $0.0, F1
    FMAXS F1, F0, F2                  // F2 = max(x, 0)
    FMOVS F2, (R0)                    // store result

    ADD $4, R0
    ADD $4, R1
    SUB $1, R3
    CBNZ R3, relu32_neon_scalar_loop

relu32_neon_done:
    RET


// func addScaledNEON(dst []float32, alpha float32, s []float32)
// dst[i] += alpha * s[i] - AXPY operation
// Frame: dst(24) + alpha(4) + pad(4) + s(24) = 56 bytes
TEXT ·addScaledNEON(SB), NOSPLIT, $0-56
    MOVD dst_base+0(FP), R0
    MOVD dst_len+8(FP), R2
    FMOVS alpha+24(FP), F3
    MOVD s_base+32(FP), R1

    // Broadcast alpha to all 4 lanes
    WORD $0x4E040463           // DUP V3.4S, V3.S[0]

    LSR $2, R2, R3
    CBZ R3, addscaled32_scalar

addscaled32_loop4:
    VLD1 (R0), [V0.S4]         // Load dst[i:i+4]
    VLD1.P 16(R1), [V1.S4]     // Load s[i:i+4]
    WORD $0x4E23CC20           // FMLA V0.4S, V1.4S, V3.4S  (dst += s * alpha)
    VST1.P [V0.S4], 16(R0)
    SUB $1, R3
    CBNZ R3, addscaled32_loop4

addscaled32_scalar:
    AND $3, R2
    CBZ R2, addscaled32_done

addscaled32_loop1:
    FMOVS (R0), F0
    FMOVS (R1), F1
    FMADDS F3, F0, F1, F0      // dst += s * alpha (Go: Fm, Fa, Fn, Fd -> Fd = Fn*Fm + Fa)
    FMOVS F0, (R0)
    ADD $4, R0
    ADD $4, R1
    SUB $1, R2
    CBNZ R2, addscaled32_loop1

addscaled32_done:
    RET
