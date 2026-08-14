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
// A representação/serializer permanece idêntica ao caminho validado
// fisicamente. A qualidade é melhorada apenas na escolha dos endpoints:
//
//  1. min/max por componente é mantido como baseline seguro;
//  2. os índices da paleta são calculados;
//  3. com esses índices fixos, novos endpoints são ajustados por
//     mínimos quadrados;
//  4. índices e endpoints são refinados por poucas iterações;
//  5. o resultado só substitui o baseline se reduzir o erro.
//
// Assim a otimização nunca piora TotalError em relação ao encoder anterior.
func EncodeMode0Block(
	pixels [16]RGBA,
) ([16]byte, uint64, error) {
	var out [16]byte

	ep0, ep1, indices, totalError :=
		optimizeMode0Endpoints(pixels)

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

func optimizeMode0Endpoints(
	pixels [16]RGBA,
) (
	RGBA,
	RGBA,
	[16]uint8,
	uint64,
) {
	ep0, ep1 := minMaxMode0Endpoints(pixels)

	indices, baselineError :=
		chooseIndices(
			pixels,
			ep0,
			ep1,
		)

	bestEP0 := ep0
	bestEP1 := ep1
	bestIndices := indices
	bestError := baselineError

	// Blocos já representados sem erro (incluindo cores sólidas)
	// precisam continuar byte por byte no caminho comprovado.
	if baselineError == 0 {
		return bestEP0,
			bestEP1,
			bestIndices,
			bestError
	}

	currentEP0 := ep0
	currentEP1 := ep1
	currentIndices := indices

	// Três passos são suficientes para corrigir a principal deficiência
	// do min/max por componente sem transformar o encoder em uma busca
	// pesada, especialmente importante para GIFs de 30 frames.
	const refinementRounds = 3

	for round := 0; round < refinementRounds; round++ {
		fittedEP0,
			fittedEP1,
			ok :=
			fitMode0Endpoints(
				pixels,
				currentIndices,
			)

		if !ok {
			break
		}

		fittedIndices,
			fittedError :=
			chooseIndices(
				pixels,
				fittedEP0,
				fittedEP1,
			)

		if fittedError < bestError {
			bestEP0 = fittedEP0
			bestEP1 = fittedEP1
			bestIndices = fittedIndices
			bestError = fittedError
		}

		if fittedEP0 == currentEP0 &&
			fittedEP1 == currentEP1 {

			break
		}

		currentEP0 = fittedEP0
		currentEP1 = fittedEP1
		currentIndices = fittedIndices
	}

	return bestEP0,
		bestEP1,
		bestIndices,
		bestError
}

func minMaxMode0Endpoints(
	pixels [16]RGBA,
) (RGBA, RGBA) {
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

	return ep0, ep1
}

// fitMode0Endpoints resolve, canal por canal, os dois endpoints que
// minimizam o erro quadrático para os índices atualmente escolhidos.
//
// A paleta usa:
//
//	value = ep0*(16-t)/16 + ep1*t/16
//
// portanto o problema possui somente duas incógnitas por canal.
func fitMode0Endpoints(
	pixels [16]RGBA,
	indices [16]uint8,
) (RGBA, RGBA, bool) {
	var sumAA float64
	var sumAB float64
	var sumBB float64

	var rhsA [4]float64
	var rhsB [4]float64

	for i, index := range indices {
		t := float64(
			mode0T[index],
		)

		a := 16.0 - t
		b := t

		sumAA += a * a
		sumAB += a * b
		sumBB += b * b

		channels := [4]uint8{
			pixels[i].R,
			pixels[i].G,
			pixels[i].B,
			pixels[i].A,
		}

		for channel, value := range channels {
			y := 16.0 *
				float64(value)

			rhsA[channel] += a * y
			rhsB[channel] += b * y
		}
	}

	determinant :=
		sumAA*sumBB -
			sumAB*sumAB

	// Todos os pixels caíram efetivamente no mesmo ponto da linha.
	// Não existe solução única; nesse caso o baseline permanece seguro.
	if determinant > -1e-9 &&
		determinant < 1e-9 {

		return RGBA{}, RGBA{}, false
	}

	var endpoint0 [4]uint8
	var endpoint1 [4]uint8

	for channel := 0; channel < 4; channel++ {
		value0 :=
			(rhsA[channel]*sumBB -
				rhsB[channel]*sumAB) /
				determinant

		value1 :=
			(rhsB[channel]*sumAA -
				rhsA[channel]*sumAB) /
				determinant

		endpoint0[channel] =
			clampMode0Channel(value0)

		endpoint1[channel] =
			clampMode0Channel(value1)
	}

	return RGBA{
			R: endpoint0[0],
			G: endpoint0[1],
			B: endpoint0[2],
			A: endpoint0[3],
		},
		RGBA{
			R: endpoint1[0],
			G: endpoint1[1],
			B: endpoint1[2],
			A: endpoint1[3],
		},
		true
}

func clampMode0Channel(
	value float64,
) uint8 {
	if value <= 0 {
		return 0
	}

	if value >= 255 {
		return 255
	}

	return uint8(
		value + 0.5,
	)
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
