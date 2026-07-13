//go:build amd64

#include "textflag.h"

// Constants
DATA absf32mask<>+0x00(SB)/4, $0x7fffffff
DATA absf32mask<>+0x04(SB)/4, $0x7fffffff
DATA absf32mask<>+0x08(SB)/4, $0x7fffffff
DATA absf32mask<>+0x0c(SB)/4, $0x7fffffff
DATA absf32mask<>+0x10(SB)/4, $0x7fffffff
DATA absf32mask<>+0x14(SB)/4, $0x7fffffff
DATA absf32mask<>+0x18(SB)/4, $0x7fffffff
DATA absf32mask<>+0x1c(SB)/4, $0x7fffffff
GLOBL absf32mask<>(SB), RODATA|NOPTR, $32

// func dotProductAVX(a, b []float32) float32
// Optimized with 4 independent accumulators to hide FMA latency.
// Processes 32 float32s per iteration (4 vectors × 8 floats).
// Handles mismatched slice lengths: uses min(len(a), len(b)).
TEXT ·dotProductAVX(SB), NOSPLIT, $0-52
    MOVQ a_base+0(FP), SI
    MOVQ a_len+8(FP), CX
    MOVQ b_len+32(FP), AX
    CMPQ AX, CX
    CMOVQLT AX, CX             // CX = min(len(a), len(b))
    MOVQ b_base+24(FP), DI

    // Initialize 4 independent accumulators
    VXORPS Y0, Y0, Y0          // acc0
    VXORPS Y3, Y3, Y3          // acc1
    VXORPS Y4, Y4, Y4          // acc2
    VXORPS Y5, Y5, Y5          // acc3

    // Process 32 elements per iteration (4 vectors × 8 floats)
    MOVQ CX, AX
    SHRQ $5, AX                // len / 32
    JZ   dot32_loop8_check

dot32_loop32:
    // Load and FMA for acc0
    VMOVUPS (SI), Y1
    VMOVUPS (DI), Y2
    VFMADD231PS Y1, Y2, Y0

    // Load and FMA for acc1
    VMOVUPS 32(SI), Y1
    VMOVUPS 32(DI), Y2
    VFMADD231PS Y1, Y2, Y3

    // Load and FMA for acc2
    VMOVUPS 64(SI), Y1
    VMOVUPS 64(DI), Y2
    VFMADD231PS Y1, Y2, Y4

    // Load and FMA for acc3
    VMOVUPS 96(SI), Y1
    VMOVUPS 96(DI), Y2
    VFMADD231PS Y1, Y2, Y5

    ADDQ $128, SI
    ADDQ $128, DI
    DECQ AX
    JNZ  dot32_loop32

    // Combine accumulators: Y0 = Y0 + Y3 + Y4 + Y5
    VADDPS Y3, Y0, Y0
    VADDPS Y4, Y0, Y0
    VADDPS Y5, Y0, Y0

dot32_loop8_check:
    // Handle remaining 8-element chunks
    ANDQ $31, CX
    MOVQ CX, AX
    SHRQ $3, AX
    JZ   dot32_remainder

dot32_loop8:
    VMOVUPS (SI), Y1
    VMOVUPS (DI), Y2
    VFMADD231PS Y1, Y2, Y0
    ADDQ $32, SI
    ADDQ $32, DI
    DECQ AX
    JNZ  dot32_loop8

dot32_remainder:
    // Reduce Y0 to X0 BEFORE scalar ops (VEX scalar ops zero upper YMM)
    VEXTRACTF128 $1, Y0, X1
    VADDPS X1, X0, X0
    VHADDPS X0, X0, X0
    VHADDPS X0, X0, X0

    ANDQ $7, CX
    JZ   dot32_done

dot32_scalar:
    VMOVSS (SI), X1
    VMOVSS (DI), X2
    VFMADD231SS X1, X2, X0
    ADDQ $4, SI
    ADDQ $4, DI
    DECQ CX
    JNZ  dot32_scalar

dot32_done:
    VMOVSS X0, ret+48(FP)
    VZEROUPPER
    RET



// func addScaledAVX(dst []float32, alpha float32, s []float32)
// dst[i] += alpha * s[i]
// Frame: dst(24) + alpha(4 padded to 8) + s(24) = 56 bytes
TEXT ·addScaledAVX(SB), NOSPLIT, $0-56
    MOVQ dst_base+0(FP), DX
    MOVQ dst_len+8(FP), CX
    MOVSS alpha+24(FP), X3
    MOVQ s_base+32(FP), SI

    // Broadcast alpha to Y3
    VBROADCASTSS X3, Y3

    MOVQ CX, AX
    SHRQ $3, AX
    JZ   addscaled32_avx_remainder

addscaled32_avx_loop8:
    VMOVUPS (SI), Y0          // s[i:i+8]
    VMOVUPS (DX), Y1          // dst[i:i+8]
    VFMADD231PS Y0, Y3, Y1    // dst += alpha * s
    VMOVUPS Y1, (DX)
    ADDQ $32, SI
    ADDQ $32, DX
    DECQ AX
    JNZ  addscaled32_avx_loop8

addscaled32_avx_remainder:
    ANDQ $7, CX
    JZ   addscaled32_avx_done

addscaled32_avx_scalar:
    VMOVSS (SI), X0
    VMOVSS (DX), X1
    VFMADD231SS X0, X3, X1
    VMOVSS X1, (DX)
    ADDQ $4, SI
    ADDQ $4, DX
    DECQ CX
    JNZ  addscaled32_avx_scalar

addscaled32_avx_done:
    VZEROUPPER
    RET


// func addScaledAVX512(dst []float32, alpha float32, s []float32)
TEXT ·addScaledAVX512(SB), NOSPLIT, $0-56
    MOVQ dst_base+0(FP), DX
    MOVQ dst_len+8(FP), CX
    MOVSS alpha+24(FP), X3
    MOVQ s_base+32(FP), SI

    // Broadcast alpha to Z3
    VBROADCASTSS X3, Z3

    MOVQ CX, AX
    SHRQ $4, AX
    JZ   addscaled32_avx512_remainder

addscaled32_avx512_loop16:
    VMOVUPS (SI), Z0          // s[i:i+16]
    VMOVUPS (DX), Z1          // dst[i:i+16]
    VFMADD231PS Z0, Z3, Z1    // dst += alpha * s
    VMOVUPS Z1, (DX)
    ADDQ $64, SI
    ADDQ $64, DX
    DECQ AX
    JNZ  addscaled32_avx512_loop16

addscaled32_avx512_remainder:
    ANDQ $15, CX
    JZ   addscaled32_avx512_done

addscaled32_avx512_scalar:
    VMOVSS (SI), X0
    VMOVSS (DX), X1
    VFMADD231SS X0, X3, X1
    VMOVSS X1, (DX)
    ADDQ $4, SI
    ADDQ $4, DX
    DECQ CX
    JNZ  addscaled32_avx512_scalar

addscaled32_avx512_done:
    VZEROUPPER
    RET



// ============================================================================
// AVX-512 implementations (16x float32 per iteration)
// ============================================================================

// func dotProductAVX512(a, b []float32) float32
// Optimized with 4 independent accumulators to hide FMA latency.
// Processes 64 float32s per iteration (4 vectors × 16 floats).
// Handles mismatched slice lengths: uses min(len(a), len(b)).
TEXT ·dotProductAVX512(SB), NOSPLIT, $0-52
    MOVQ a_base+0(FP), SI
    MOVQ a_len+8(FP), CX
    MOVQ b_len+32(FP), AX
    CMPQ AX, CX
    CMOVQLT AX, CX             // CX = min(len(a), len(b))
    MOVQ b_base+24(FP), DI

    // Initialize 4 independent accumulators
    VXORPS Z0, Z0, Z0          // acc0
    VXORPS Z3, Z3, Z3          // acc1
    VXORPS Z4, Z4, Z4          // acc2
    VXORPS Z5, Z5, Z5          // acc3

    // Process 64 elements per iteration (4 vectors × 16 floats)
    MOVQ CX, AX
    SHRQ $6, AX                // len / 64
    JZ   dot32_512_loop16_check

dot32_512_loop64:
    // Load and FMA for acc0
    VMOVUPS (SI), Z1
    VMOVUPS (DI), Z2
    VFMADD231PS Z1, Z2, Z0

    // Load and FMA for acc1
    VMOVUPS 64(SI), Z1
    VMOVUPS 64(DI), Z2
    VFMADD231PS Z1, Z2, Z3

    // Load and FMA for acc2
    VMOVUPS 128(SI), Z1
    VMOVUPS 128(DI), Z2
    VFMADD231PS Z1, Z2, Z4

    // Load and FMA for acc3
    VMOVUPS 192(SI), Z1
    VMOVUPS 192(DI), Z2
    VFMADD231PS Z1, Z2, Z5

    ADDQ $256, SI
    ADDQ $256, DI
    DECQ AX
    JNZ  dot32_512_loop64

    // Combine accumulators: Z0 = Z0 + Z3 + Z4 + Z5
    VADDPS Z3, Z0, Z0
    VADDPS Z4, Z0, Z0
    VADDPS Z5, Z0, Z0

dot32_512_loop16_check:
    // Handle remaining 16-element chunks
    ANDQ $63, CX
    MOVQ CX, AX
    SHRQ $4, AX
    JZ   dot32_512_remainder

dot32_512_loop16:
    VMOVUPS (SI), Z1
    VMOVUPS (DI), Z2
    VFMADD231PS Z1, Z2, Z0
    ADDQ $64, SI
    ADDQ $64, DI
    DECQ AX
    JNZ  dot32_512_loop16

dot32_512_remainder:
    VEXTRACTF32X8 $1, Z0, Y1
    VADDPS Y1, Y0, Y0
    VEXTRACTF128 $1, Y0, X1
    VADDPS X1, X0, X0
    VHADDPS X0, X0, X0
    VHADDPS X0, X0, X0

    ANDQ $15, CX
    JZ   dot32_512_done

dot32_512_scalar:
    VMOVSS (SI), X1
    VMOVSS (DI), X2
    VFMADD231SS X1, X2, X0
    ADDQ $4, SI
    ADDQ $4, DI
    DECQ CX
    JNZ  dot32_512_scalar

dot32_512_done:
    VMOVSS X0, ret+48(FP)
    VZEROUPPER
    RET



// func reluAVX(dst, src []float32)
// Computes ReLU: dst[i] = max(0, src[i])
TEXT ·reluAVX(SB), NOSPLIT, $0-48
    MOVQ dst_base+0(FP), DX
    MOVQ dst_len+8(FP), CX
    MOVQ src_base+24(FP), SI

    // Create zero vector
    VXORPS Y1, Y1, Y1                  // Y1 = 0

    // Process 8 elements per iteration
    MOVQ CX, AX
    SHRQ $3, AX                        // len / 8
    JZ   relu32_remainder

relu32_loop8:
    VMOVUPS (SI), Y0                   // Y0 = src[i]
    VMAXPS Y0, Y1, Y0                  // Y0 = max(src[i], 0)
    VMOVUPS Y0, (DX)                   // store result
    ADDQ $32, SI
    ADDQ $32, DX
    DECQ AX
    JNZ  relu32_loop8

relu32_remainder:
    ANDQ $7, CX                        // remainder = len % 8
    JZ   relu32_done
    VXORPS X1, X1, X1                  // X1 = 0 (scalar)

relu32_scalar:
    VMOVSS (SI), X0                    // X0 = src[i]
    VMAXSS X0, X1, X0                  // X0 = max(src[i], 0)
    VMOVSS X0, (DX)                    // store result
    ADDQ $4, SI
    ADDQ $4, DX
    DECQ CX
    JNZ  relu32_scalar

relu32_done:
    VZEROUPPER
    RET
