package stc

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestEncodeImageOpaqueSize(t *testing.T) {
	img := image.NewRGBA(
		image.Rect(
			0,
			0,
			240,
			240,
		),
	)

	fillImage(
		img,
		color.RGBA{
			R: 0,
			G: 0,
			B: 80,
			A: 255,
		},
	)

	frame, stats, err := EncodeImageOpaque(img)
	if err != nil {
		t.Fatal(err)
	}

	if len(frame) != FrameSize {
		t.Fatalf(
			"frame=%d esperado=%d",
			len(frame),
			FrameSize,
		)
	}

	if FrameSize != 0x9000 {
		t.Fatalf(
			"FrameSize=0x%X esperado=0x9000",
			FrameSize,
		)
	}

	if stats.Blocks != 2304 {
		t.Fatalf(
			"blocks=%d esperado=2304",
			stats.Blocks,
		)
	}

	if stats.Pixels != 192*192 {
		t.Fatalf(
			"pixels=%d esperado=%d",
			stats.Pixels,
			192*192,
		)
	}

	if stats.TotalError != 0 {
		t.Fatalf(
			"erro esperado=0 obtido=%d",
			stats.TotalError,
		)
	}
}

func TestEncodeImageOpaqueVendorBlue(t *testing.T) {
	img := image.NewRGBA(
		image.Rect(
			0,
			0,
			192,
			192,
		),
	)

	fillImage(
		img,
		color.RGBA{
			R: 0,
			G: 0,
			B: 80,
			A: 255,
		},
	)

	frame, _, err := EncodeImageOpaque(img)
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte{
		0x00, 0x00, 0x00, 0x00,
		0xA0, 0xA0, 0xFE, 0xFF,
		0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	if !bytes.Equal(frame[:16], expected) {
		t.Fatalf(
			"primeiro bloco:\nesperado=%x\nobtido  =%x",
			expected,
			frame[:16],
		)
	}

	for i := 0; i < BlocksX*BlocksY; i++ {
		start := i * BlockSize
		end := start + BlockSize

		block := frame[start:end]

		if !bytes.Equal(block, expected) {
			t.Fatalf(
				"bloco %d diferente: %x",
				i,
				block,
			)
		}
	}
}

func TestEncodeImageOpaqueForcesAlpha(t *testing.T) {
	img := image.NewRGBA(
		image.Rect(
			0,
			0,
			4,
			4,
		),
	)

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(
				x,
				y,
				color.RGBA{
					R: 180,
					G: 0,
					B: 0,
					A: 0,
				},
			)
		}
	}

	frame, stats, err := EncodeImageOpaque(img)
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalError != 0 {
		t.Fatalf(
			"erro esperado=0 obtido=%d",
			stats.TotalError,
		)
	}

	expected := []byte{
		0x68, 0x69, 0x01, 0x00,
		0x00, 0x00, 0xFE, 0xFF,
		0x01, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	if !bytes.Equal(frame[:16], expected) {
		t.Fatalf(
			"esperado vermelho opaco=%x\nobtido=%x",
			expected,
			frame[:16],
		)
	}
}

func TestEncodeImageOpaqueDifferentSize(t *testing.T) {
	img := image.NewRGBA(
		image.Rect(
			0,
			0,
			320,
			180,
		),
	)

	fillImage(
		img,
		color.RGBA{
			R: 30,
			G: 90,
			B: 150,
			A: 255,
		},
	)

	frame, stats, err := EncodeImageOpaque(img)
	if err != nil {
		t.Fatal(err)
	}

	if len(frame) != 0x9000 {
		t.Fatalf(
			"frame=0x%X esperado=0x9000",
			len(frame),
		)
	}

	if stats.Blocks != 2304 {
		t.Fatalf(
			"blocos=%d esperado=2304",
			stats.Blocks,
		)
	}
}

func fillImage(
	img *image.RGBA,
	c color.RGBA,
) {
	b := img.Bounds()

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			img.SetRGBA(
				x,
				y,
				c,
			)
		}
	}
}
