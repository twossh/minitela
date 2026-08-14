package customimage

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultBuildPaths retorna:
//
//   - o template ACF compatível disponibilizado pelo MiniTela;
//   - o Texture-custom.acf gerado pelo MiniTela.
//
// No AppImage, o template é interno e invisível ao usuário.
// O template nunca é alterado.
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

	templatePath, err =
		ResolveTemplatePath()
	if err != nil {
		return "", "", err
	}

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
