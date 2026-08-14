package customimage

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultBuildPaths retorna:
//
//   - o Texture1.acf oficial usado apenas como template;
//   - o Texture-custom.acf gerado pelo MiniTela.
//
// O arquivo oficial nunca é alterado.
func DefaultBuildPaths() (
	templatePath string,
	outputPath string,
	err error,
) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf(
			"localizar HOME: %w",
			err,
		)
	}

	templatePath =
		filepath.Join(
			home,
			".local",
			"share",
			"minitela",
			"vendor",
			"textures",
			"Texture1.acf",
		)

	outputPath =
		filepath.Join(
			home,
			".local",
			"share",
			"minitela",
			"custom",
			"Texture-custom.acf",
		)

	return templatePath,
		outputPath,
		nil
}
