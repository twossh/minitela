package stc

import (
	"fmt"
	"image"
)

const (
	ImageWidth  = 192
	ImageHeight = 192

	BlockWidth  = 4
	BlockHeight = 4

	BlocksX = ImageWidth / BlockWidth
	BlocksY = ImageHeight / BlockHeight

	BlockSize = 16
	FrameSize = BlocksX * BlocksY * BlockSize
)

// ImageStats contém estatísticas da compressão
// de uma imagem completa.
type ImageStats struct {
	Blocks     int
	Pixels     int
	TotalError uint64
}

// EncodeImageOpaque converte qualquer image.Image em um
// frame STCRGBA 192x192 de exatamente 0x9000 bytes.
//
// Nesta primeira implementação o alpha é sempre 255,
// pois esse caminho já foi validado fisicamente no GC9002.
func EncodeImageOpaque(src image.Image) ([]byte, ImageStats, error) {
	var stats ImageStats

	if src == nil {
		return nil, stats, fmt.Errorf("STCRGBA: imagem nil")
	}

	resized := resizeBilinear(
		src,
		ImageWidth,
		ImageHeight,
	)

	frame := make([]byte, 0, FrameSize)

	for by := 0; by < BlocksY; by++ {
		for bx := 0; bx < BlocksX; bx++ {
			var pixels [16]RGBA

			k := 0

			for y := 0; y < BlockHeight; y++ {
				for x := 0; x < BlockWidth; x++ {
					px := bx*BlockWidth + x
					py := by*BlockHeight + y

					pos := py*ImageWidth + px
					c := resized[pos]

					pixels[k] = RGBA{
						R: c.R,
						G: c.G,
						B: c.B,
						A: 255,
					}

					k++
				}
			}

			raw, blockError, err := EncodeMode0Block(pixels)
			if err != nil {
				return nil, stats, fmt.Errorf(
					"STCRGBA: bloco (%d,%d): %w",
					bx,
					by,
					err,
				)
			}

			frame = append(frame, raw[:]...)

			stats.Blocks++
			stats.TotalError += blockError
		}
	}

	stats.Pixels = ImageWidth * ImageHeight

	if len(frame) != FrameSize {
		return nil, stats, fmt.Errorf(
			"STCRGBA: frame possui 0x%X bytes; esperado 0x%X",
			len(frame),
			FrameSize,
		)
	}

	return frame, stats, nil
}

type rgbaPixel struct {
	R uint8
	G uint8
	B uint8
	A uint8
}

// resizeBilinear redimensiona a imagem sem depender
// de bibliotecas externas.
func resizeBilinear(
	src image.Image,
	width int,
	height int,
) []rgbaPixel {
	dst := make([]rgbaPixel, width*height)

	bounds := src.Bounds()

	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= 0 || srcHeight <= 0 {
		return dst
	}

	if srcWidth == 1 || srcHeight == 1 {
		return resizeNearest(
			src,
			width,
			height,
		)
	}

	for dy := 0; dy < height; dy++ {
		srcY := (float64(dy)+0.5)*
			float64(srcHeight)/
			float64(height) - 0.5

		y0 := floorInt(srcY)
		fy := srcY - float64(y0)

		if y0 < 0 {
			y0 = 0
			fy = 0
		}

		y1 := y0 + 1

		if y1 >= srcHeight {
			y1 = srcHeight - 1
		}

		for dx := 0; dx < width; dx++ {
			srcX := (float64(dx)+0.5)*
				float64(srcWidth)/
				float64(width) - 0.5

			x0 := floorInt(srcX)
			fx := srcX - float64(x0)

			if x0 < 0 {
				x0 = 0
				fx = 0
			}

			x1 := x0 + 1

			if x1 >= srcWidth {
				x1 = srcWidth - 1
			}

			c00 := colorAt(
				src,
				bounds.Min.X+x0,
				bounds.Min.Y+y0,
			)

			c10 := colorAt(
				src,
				bounds.Min.X+x1,
				bounds.Min.Y+y0,
			)

			c01 := colorAt(
				src,
				bounds.Min.X+x0,
				bounds.Min.Y+y1,
			)

			c11 := colorAt(
				src,
				bounds.Min.X+x1,
				bounds.Min.Y+y1,
			)

			pos := dy*width + dx

			dst[pos] = rgbaPixel{
				R: bilinearChannel(
					c00.R,
					c10.R,
					c01.R,
					c11.R,
					fx,
					fy,
				),
				G: bilinearChannel(
					c00.G,
					c10.G,
					c01.G,
					c11.G,
					fx,
					fy,
				),
				B: bilinearChannel(
					c00.B,
					c10.B,
					c01.B,
					c11.B,
					fx,
					fy,
				),
				A: bilinearChannel(
					c00.A,
					c10.A,
					c01.A,
					c11.A,
					fx,
					fy,
				),
			}
		}
	}

	return dst
}

func resizeNearest(
	src image.Image,
	width int,
	height int,
) []rgbaPixel {
	dst := make([]rgbaPixel, width*height)

	bounds := src.Bounds()

	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	if srcWidth <= 0 || srcHeight <= 0 {
		return dst
	}

	for y := 0; y < height; y++ {
		sy := y * srcHeight / height

		if sy >= srcHeight {
			sy = srcHeight - 1
		}

		for x := 0; x < width; x++ {
			sx := x * srcWidth / width

			if sx >= srcWidth {
				sx = srcWidth - 1
			}

			pos := y*width + x

			dst[pos] = colorAt(
				src,
				bounds.Min.X+sx,
				bounds.Min.Y+sy,
			)
		}
	}

	return dst
}

func colorAt(
	src image.Image,
	x int,
	y int,
) rgbaPixel {
	r, g, b, a := src.At(x, y).RGBA()

	return rgbaPixel{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

func bilinearChannel(
	c00 uint8,
	c10 uint8,
	c01 uint8,
	c11 uint8,
	fx float64,
	fy float64,
) uint8 {
	top := float64(c00)*(1.0-fx) +
		float64(c10)*fx

	bottom := float64(c01)*(1.0-fx) +
		float64(c11)*fx

	value := top*(1.0-fy) +
		bottom*fy

	if value <= 0 {
		return 0
	}

	if value >= 255 {
		return 255
	}

	return uint8(value + 0.5)
}

func floorInt(v float64) int {
	i := int(v)

	if float64(i) > v {
		return i - 1
	}

	return i
}
