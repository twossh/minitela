package customimage

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/twossh/minitela/internal/acf"
	"github.com/twossh/minitela/internal/stc"
)

// BuildStats descreve o resultado da geração de uma
// textura personalizada.
type BuildStats struct {
	Format       string
	SourceWidth  int
	SourceHeight int
	OutputSize   int
	Checksum     uint32
	STC          stc.ImageStats
}

// BuildStaticTextureFile converte uma imagem comum em
// um ACF estático compatível com a página de imagem do R15M.
//
// Formatos registrados:
//   - PNG
//   - JPEG
//   - GIF
//
// Para GIF, esta versão estática utiliza o primeiro frame.
// O suporte animado será implementado separadamente.
func BuildStaticTextureFile(
	inputPath string,
	templatePath string,
	outputPath string,
) (BuildStats, error) {
	var stats BuildStats

	if inputPath == "" {
		return stats, fmt.Errorf(
			"arquivo de imagem não informado",
		)
	}

	if templatePath == "" {
		return stats, fmt.Errorf(
			"template ACF não informado",
		)
	}

	if outputPath == "" {
		return stats, fmt.Errorf(
			"arquivo ACF de saída não informado",
		)
	}

	input, err := os.Open(inputPath)
	if err != nil {
		return stats, fmt.Errorf(
			"abrir imagem %q: %w",
			inputPath,
			err,
		)
	}
	defer input.Close()

	img, format, err := image.Decode(input)
	if err != nil {
		return stats, fmt.Errorf(
			"decodificar imagem %q: %w",
			inputPath,
			err,
		)
	}

	bounds := img.Bounds()

	stats.Format = format
	stats.SourceWidth = bounds.Dx()
	stats.SourceHeight = bounds.Dy()

	if stats.SourceWidth <= 0 ||
		stats.SourceHeight <= 0 {
		return stats, fmt.Errorf(
			"imagem inválida: %dx%d",
			stats.SourceWidth,
			stats.SourceHeight,
		)
	}

	frame, stcStats, err :=
		stc.EncodeImageOpaque(img)

	if err != nil {
		return stats, fmt.Errorf(
			"gerar STCRGBA: %w",
			err,
		)
	}

	stats.STC = stcStats

	template, err := os.ReadFile(
		templatePath,
	)
	if err != nil {
		return stats, fmt.Errorf(
			"ler template ACF %q: %w",
			templatePath,
			err,
		)
	}

	if err := acf.ValidateTextureTemplate(
		template,
	); err != nil {
		return stats, fmt.Errorf(
			"template ACF: %w",
			err,
		)
	}

	result, err := acf.BuildStaticTexture(
		template,
		frame,
	)
	if err != nil {
		return stats, fmt.Errorf(
			"montar ACF: %w",
			err,
		)
	}

	if err := acf.ValidateChecksum(
		result,
	); err != nil {
		return stats, fmt.Errorf(
			"validar ACF final: %w",
			err,
		)
	}

	checksum, err := acf.StoredChecksum(
		result,
	)
	if err != nil {
		return stats, fmt.Errorf(
			"ler checksum final: %w",
			err,
		)
	}

	stats.Checksum = checksum
	stats.OutputSize = len(result)

	outputDir := filepath.Dir(
		outputPath,
	)

	if err := os.MkdirAll(
		outputDir,
		0o755,
	); err != nil {
		return stats, fmt.Errorf(
			"criar diretório %q: %w",
			outputDir,
			err,
		)
	}

	tmpPath := outputPath + ".tmp"

	if err := os.WriteFile(
		tmpPath,
		result,
		0o644,
	); err != nil {
		return stats, fmt.Errorf(
			"gravar ACF temporário: %w",
			err,
		)
	}

	if err := os.Rename(
		tmpPath,
		outputPath,
	); err != nil {
		_ = os.Remove(tmpPath)

		return stats, fmt.Errorf(
			"finalizar ACF %q: %w",
			outputPath,
			err,
		)
	}

	return stats, nil
}
