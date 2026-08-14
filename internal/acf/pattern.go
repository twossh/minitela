package acf

import (
	"image/color"
)

func QuadrantPattern(
	packing VendorPacking,
	swapRB bool,
) []byte {
	result :=
		make(
			[]byte,
			BC3FrameSize,
		)

	red :=
		color.RGBA{
			R: 255,
			A: 255,
		}

	green :=
		color.RGBA{
			G: 255,
			A: 255,
		}

	blue :=
		color.RGBA{
			B: 255,
			A: 255,
		}

	white :=
		color.RGBA{
			R: 255,
			G: 255,
			B: 255,
			A: 255,
		}

	for by := 0; by < BC3BlocksY; by++ {

		for bx := 0; bx < BC3BlocksX; bx++ {

			var c color.RGBA

			switch {
			case bx < BC3BlocksX/2 &&
				by < BC3BlocksY/2:

				c = red

			case bx >= BC3BlocksX/2 &&
				by < BC3BlocksY/2:

				c = green

			case bx < BC3BlocksX/2 &&
				by >= BC3BlocksY/2:

				c = blue

			default:
				c = white
			}

			block :=
				SolidBC3Block(
					c,
					packing,
					swapRB,
				)

			index :=
				by*BC3BlocksX +
					bx

			start :=
				index *
					BC3BlockSize

			end := start + BC3BlockSize

			copy(
				result[start:end],
				block[:],
			)
		}
	}

	return result
}
