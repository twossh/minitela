package customimage

import (
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/twossh/minitela/internal/acf"
)

func TestBuildStaticTextureFile(
	t *testing.T,
) {
	dir := t.TempDir()

	imagePath := filepath.Join(
		dir,
		"input.png",
	)

	templatePath := filepath.Join(
		dir,
		"template.acf",
	)

	outputPath := filepath.Join(
		dir,
		"output.acf",
	)

	img := image.NewRGBA(
		image.Rect(
			0,
			0,
			320,
			180,
		),
	)

	for y := 0; y < 180; y++ {
		for x := 0; x < 320; x++ {
			img.SetRGBA(
				x,
				y,
				color.RGBA{
					R: uint8(
						x * 255 / 319,
					),
					G: uint8(
						y * 255 / 179,
					),
					B: 80,
					A: 255,
				},
			)
		}
	}

	f, err := os.Create(
		imagePath,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := png.Encode(
		f,
		img,
	); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}

	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	template := make(
		[]byte,
		acf.TextureTemplateSize,
	)

	for i := range template {
		template[i] = byte(
			(i*17 + 3) & 0xFF,
		)
	}

	binary.LittleEndian.PutUint32(
		template[len(template)-4:],
		acf.FooterMagic,
	)

	if err := acf.SetChecksum(
		template,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		templatePath,
		template,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	stats, err := BuildStaticTextureFile(
		imagePath,
		templatePath,
		outputPath,
	)
	if err != nil {
		t.Fatal(err)
	}

	if stats.Format != "png" {
		t.Fatalf(
			"formato=%q esperado=png",
			stats.Format,
		)
	}

	if stats.SourceWidth != 320 ||
		stats.SourceHeight != 180 {
		t.Fatalf(
			"dimensão=%dx%d",
			stats.SourceWidth,
			stats.SourceHeight,
		)
	}

	if stats.STC.Blocks != 2304 {
		t.Fatalf(
			"blocos=%d esperado=2304",
			stats.STC.Blocks,
		)
	}

	if stats.OutputSize !=
		acf.TextureTemplateSize {
		t.Fatalf(
			"saída=0x%X esperado=0x%X",
			stats.OutputSize,
			acf.TextureTemplateSize,
		)
	}

	result, err := os.ReadFile(
		outputPath,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := acf.ValidateChecksum(
		result,
	); err != nil {
		t.Fatalf(
			"ACF final inválido: %v",
			err,
		)
	}

	stored, err := acf.StoredChecksum(
		result,
	)
	if err != nil {
		t.Fatal(err)
	}

	if stored != stats.Checksum {
		t.Fatalf(
			"checksum stats=0x%08X arquivo=0x%08X",
			stats.Checksum,
			stored,
		)
	}
}
