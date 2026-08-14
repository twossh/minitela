package stc

import "fmt"

// RGBA representa um pixel usado pelo compressor STCRGBA.
type RGBA struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

var mode0T = [16]int{
	0, 1, 2, 3,
	4, 5, 6, 7,
	9, 10, 11, 12,
	13, 14, 15, 16,
}

// EncodeMode0Block comprime exatamente 16 pixels (um bloco 4x4)
// usando o formato STCRGBA Mode 0 validado fisicamente no GC9002.
//
// Esta primeira implementação utiliza endpoints min/max por componente.
// É válida, porém ainda não reproduz a otimização de qualidade completa
// do gerador original.
func EncodeMode0Block(
	pixels [16]RGBA,
) ([16]byte, uint64, error) {
	var out [16]byte

	ep0 := RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	}

	ep1 := RGBA{}

	for _, p := range pixels {
		if p.R < ep0.R {
			ep0.R = p.R
		}
		if p.G < ep0.G {
			ep0.G = p.G
		}
		if p.B < ep0.B {
			ep0.B = p.B
		}
		if p.A < ep0.A {
			ep0.A = p.A
		}

		if p.R > ep1.R {
			ep1.R = p.R
		}
		if p.G > ep1.G {
			ep1.G = p.G
		}
		if p.B > ep1.B {
			ep1.B = p.B
		}
		if p.A > ep1.A {
			ep1.A = p.A
		}
	}

	indices, totalError := chooseIndices(
		pixels,
		ep0,
		ep1,
	)

	// O serializer reserva somente 3 bits para index[15].
	// O gerador original resolve isso invertendo endpoints e índices.
	if indices[15] >= 8 {
		ep0, ep1 = ep1, ep0

		for i := range indices {
			indices[i] = 15 - indices[i]
		}
	}

	if indices[15] >= 8 {
		return out, 0, fmt.Errorf(
			"STCRGBA Mode0: index[15]=%d não cabe em 3 bits",
			indices[15],
		)
	}

	bw := bitWriter{
		data: &out,
	}

	// Prefixo do Mode 0.
	bw.write(0, 1)

	// Ordem comprovada no serializer 0x4AB7F0:
	//
	// R0 R1
	// G0 G1
	// B0 B1
	// A0 A1
	bw.write(uint32(ep0.R), 8)
	bw.write(uint32(ep1.R), 8)

	bw.write(uint32(ep0.G), 8)
	bw.write(uint32(ep1.G), 8)

	bw.write(uint32(ep0.B), 8)
	bw.write(uint32(ep1.B), 8)

	bw.write(uint32(ep0.A), 8)
	bw.write(uint32(ep1.A), 8)

	// index[0..14] = 4 bits.
	for i := 0; i < 15; i++ {
		bw.write(
			uint32(indices[i]),
			4,
		)
	}

	// index[15] = 3 bits.
	bw.write(
		uint32(indices[15]),
		3,
	)

	if bw.bit != 128 {
		return out, 0, fmt.Errorf(
			"STCRGBA Mode0: bitstream terminou em %d bits",
			bw.bit,
		)
	}

	return out, totalError, nil
}

func chooseIndices(
	pixels [16]RGBA,
	ep0 RGBA,
	ep1 RGBA,
) ([16]uint8, uint64) {
	var indices [16]uint8

	palette := mode0Palette(
		ep0,
		ep1,
	)

	var total uint64

	for pixelIndex, px := range pixels {
		var bestIndex uint8
		bestError := uint64(^uint64(0))

		for paletteIndex, candidate := range palette {
			err := rgbaError(
				px,
				candidate,
			)

			if err < bestError {
				bestError = err
				bestIndex = uint8(
					paletteIndex,
				)
			}
		}

		indices[pixelIndex] = bestIndex
		total += bestError
	}

	return indices, total
}

func mode0Palette(
	ep0 RGBA,
	ep1 RGBA,
) [16]RGBA {
	var palette [16]RGBA

	for i, t := range mode0T {
		palette[i] = RGBA{
			R: interpolate(
				ep0.R,
				ep1.R,
				t,
			),
			G: interpolate(
				ep0.G,
				ep1.G,
				t,
			),
			B: interpolate(
				ep0.B,
				ep1.B,
				t,
			),
			A: interpolate(
				ep0.A,
				ep1.A,
				t,
			),
		}
	}

	return palette
}

func interpolate(
	a uint8,
	b uint8,
	t int,
) uint8 {
	// Divisão inteira em Go trunca em direção a zero,
	// igual à cvttsd2si observada no gerador original.
	value := int(a) +
		((int(b)-int(a))*t)/16

	if value < 0 {
		value = 0
	}

	if value > 255 {
		value = 255
	}

	return uint8(value)
}

func rgbaError(
	a RGBA,
	b RGBA,
) uint64 {
	dr := int64(a.R) - int64(b.R)
	dg := int64(a.G) - int64(b.G)
	db := int64(a.B) - int64(b.B)
	da := int64(a.A) - int64(b.A)

	return uint64(
		dr*dr +
			dg*dg +
			db*db +
			da*da,
	)
}

type bitWriter struct {
	data *[16]byte
	bit  int
}

func (w *bitWriter) write(
	value uint32,
	count int,
) {
	for i := 0; i < count; i++ {
		if (value>>i)&1 != 0 {
			w.data[w.bit>>3] |=
				1 << (w.bit & 7)
		}

		w.bit++
	}
}
