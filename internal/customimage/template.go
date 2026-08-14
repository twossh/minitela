package customimage

import (
	"fmt"
	"os"

	"github.com/twossh/minitela/internal/acf"
)

const internalTemplateEnv = "MINITELA_TEMPLATE"

// ResolveTemplatePath retorna o template interno disponibilizado pelo
// empacotamento do MiniTela. O usuário final não configura arquivos ACF.
func ResolveTemplatePath() (string, error) {
	path := os.Getenv(
		internalTemplateEnv,
	)

	if path == "" {
		return "", fmt.Errorf(
			"componente interno de imagem não encontrado",
		)
	}

	if err :=
		validateTemplateFile(
			path,
		); err != nil {
		return "", fmt.Errorf(
			"componente interno de imagem inválido: %w",
			err,
		)
	}

	return path, nil
}

func validateTemplateFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err :=
		acf.ValidateTextureTemplate(
			data,
		); err != nil {
		return fmt.Errorf(
			"%s: %w",
			path,
			err,
		)
	}

	return nil
}
