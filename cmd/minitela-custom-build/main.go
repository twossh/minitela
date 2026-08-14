package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/twossh/minitela/internal/acf"
	"github.com/twossh/minitela/internal/customimage"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fail(
			fmt.Errorf(
				"descobrir HOME: %w",
				err,
			),
		)
	}

	defaultOutput := filepath.Join(
		home,
		".local",
		"share",
		"minitela",
		"custom",
		"Texture-custom.acf",
	)

	input := flag.String(
		"input",
		"",
		"imagem PNG/JPG/GIF de entrada",
	)

	template := flag.String(
		"template",
		os.Getenv("MINITELA_TEMPLATE"),
		"template ACF compatível (uso de desenvolvimento)",
	)

	output := flag.String(
		"output",
		defaultOutput,
		"arquivo ACF de saída",
	)

	flag.Parse()

	if strings.TrimSpace(
		*input,
	) == "" {
		fmt.Fprintln(
			os.Stderr,
			"Erro: informe -input <imagem>",
		)

		fmt.Fprintln(
			os.Stderr,
			"Exemplo:",
		)

		fmt.Fprintln(
			os.Stderr,
			"  MiniTelaCustomBuild -input foto.png",
		)

		os.Exit(2)
	}

	fmt.Println(
		"=== MiniTela - Custom Texture Builder ===",
	)

	fmt.Println()

	fmt.Printf(
		"Imagem   : %s\n",
		*input,
	)

	fmt.Printf(
		"Template : %s\n",
		*template,
	)

	fmt.Printf(
		"Saída    : %s\n",
		*output,
	)

	fmt.Println()
	fmt.Println(
		"Gerando STCRGBA nativo...",
	)

	stats, err :=
		customimage.BuildStaticTextureFile(
			*input,
			*template,
			*output,
		)

	if err != nil {
		fail(err)
	}

	fmt.Println()
	fmt.Println(
		"=== RESULTADO ===",
	)

	fmt.Println()

	fmt.Printf(
		"Formato        : %s\n",
		stats.Format,
	)

	fmt.Printf(
		"Imagem original: %dx%d\n",
		stats.SourceWidth,
		stats.SourceHeight,
	)

	fmt.Printf(
		"Imagem R15M   : 192x192\n",
	)

	fmt.Printf(
		"Blocos STC    : %d\n",
		stats.STC.Blocks,
	)

	fmt.Printf(
		"Frame STC     : 0x%X bytes\n",
		acf.TextureFrameSize,
	)

	fmt.Printf(
		"Frames ACF    : %d\n",
		acf.TextureFrameCount,
	)

	fmt.Printf(
		"Erro SSE      : %d\n",
		stats.STC.TotalError,
	)

	fmt.Printf(
		"ACF tamanho   : 0x%X bytes\n",
		stats.OutputSize,
	)

	fmt.Printf(
		"Checksum      : 0x%08X\n",
		stats.Checksum,
	)

	data, err := os.ReadFile(
		*output,
	)
	if err != nil {
		fail(
			fmt.Errorf(
				"reabrir ACF final: %w",
				err,
			),
		)
	}

	hash := sha256.Sum256(
		data,
	)

	fmt.Printf(
		"SHA256        : %x\n",
		hash,
	)

	fmt.Println()

	fmt.Printf(
		"Arquivo       : %s\n",
		*output,
	)

	fmt.Println()
	fmt.Println(
		"STATUS        : OK",
	)
}

func fail(err error) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %v\n",
		err,
	)

	os.Exit(1)
}
