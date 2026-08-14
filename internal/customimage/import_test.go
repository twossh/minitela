package customimage

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestImportRejectsEmpty(t *testing.T) {
	_, err :=
		Import(
			bytes.NewReader(nil),
			"teste.png",
		)

	if err == nil {
		t.Fatal(
			"esperava erro para arquivo vazio",
		)
	}
}

func TestDecodePNG(t *testing.T) {
	img :=
		image.NewRGBA(
			image.Rect(
				0,
				0,
				32,
				24,
			),
		)

	img.Set(
		0,
		0,
		color.RGBA{
			R: 255,
			A: 255,
		},
	)

	var buffer bytes.Buffer

	if err :=
		png.Encode(
			&buffer,
			img,
		); err != nil {
		t.Fatal(err)
	}

	config, format, err :=
		image.DecodeConfig(
			bytes.NewReader(
				buffer.Bytes(),
			),
		)

	if err != nil {
		t.Fatal(err)
	}

	if format != "png" {
		t.Fatalf(
			"format=%q esperado=png",
			format,
		)
	}

	if config.Width != 32 ||
		config.Height != 24 {
		t.Fatalf(
			"dimensão=%dx%d esperado=32x24",
			config.Width,
			config.Height,
		)
	}

	_ = os.ErrNotExist
}
