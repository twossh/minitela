package acf

import (
	"encoding/binary"
	"fmt"
	"image/color"
)

const (
	TestWidth    = 192
	TestHeight   = 192
	BC3BlockSize = 16

	BC3BlocksX = TestWidth / 4
	BC3BlocksY = TestHeight / 4

	BC3FrameSize = BC3BlocksX *
		BC3BlocksY *
		BC3BlockSize
)

type VendorPacking int

const (
	// Vendor:
	// [COR: endpoints + selectors]
	// [ALPHA]
	PackingColorFirst VendorPacking = iota

	// BC3 padrão:
	// [ALPHA]
	// [COR]
	PackingAlphaFirst

	// Hipótese R15M:
	//
	// [selector uint32]
	// [endpoint0 uint16]
	// [endpoint1 uint16]
	// [alpha 8 bytes]
	PackingColorSelectorsFirst
)

func RGB565(
	c color.RGBA,
	swapRB bool,
) uint16 {
	r := uint16(c.R >> 3)
	g := uint16(c.G >> 2)
	b := uint16(c.B >> 3)

	if swapRB {
		r, b = b, r
	}

	return (r << 11) |
		(g << 5) |
		b
}

func SolidBC3Block(
	c color.RGBA,
	packing VendorPacking,
	swapRB bool,
) [16]byte {
	var standard [16]byte

	//
	// BC3 padrão
	//

	// Alpha opaco.
	standard[0] = 0xFF
	standard[1] = 0xFF

	// alpha selectors = zero.

	rgb := RGB565(
		c,
		swapRB,
	)

	// Para uma cor sólida usamos
	// os dois endpoints iguais.
	binary.LittleEndian.PutUint16(
		standard[8:10],
		rgb,
	)

	binary.LittleEndian.PutUint16(
		standard[10:12],
		rgb,
	)

	// Todos os pixels usam selector 0.
	binary.LittleEndian.PutUint32(
		standard[12:16],
		0,
	)

	var vendor [16]byte

	switch packing {

	case PackingColorFirst:
		//
		// [endpoint0 endpoint1 selectors]
		// [alpha]
		//

		copy(
			vendor[0:8],
			standard[8:16],
		)

		copy(
			vendor[8:16],
			standard[0:8],
		)

	case PackingAlphaFirst:
		//
		// BC3 padrão.
		//

		copy(
			vendor[:],
			standard[:],
		)

	case PackingColorSelectorsFirst:
		//
		// Hipótese obtida pelo WordMap:
		//
		// word 0 + word 1:
		// selectors 32-bit
		//
		// word 2:
		// endpoint0
		//
		// word 3:
		// endpoint1
		//

		copy(
			vendor[0:4],
			standard[12:16],
		)

		copy(
			vendor[4:6],
			standard[8:10],
		)

		copy(
			vendor[6:8],
			standard[10:12],
		)

		// Alpha permanece na segunda metade.
		copy(
			vendor[8:16],
			standard[0:8],
		)

	default:
		panic(
			fmt.Sprintf(
				"packing desconhecido: %d",
				packing,
			),
		)
	}

	return vendor
}
