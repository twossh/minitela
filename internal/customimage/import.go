package customimage

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Info struct {
	Path   string
	Name   string
	Format string

	Width  int
	Height int
	Frames int

	Size int64
}

func Import(
	reader io.Reader,
	originalName string,
) (*Info, error) {
	if reader == nil {
		return nil, fmt.Errorf(
			"imagem não informada",
		)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf(
			"ler imagem: %w",
			err,
		)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf(
			"arquivo vazio",
		)
	}

	extension :=
		strings.ToLower(
			filepath.Ext(originalName),
		)

	switch extension {
	case ".png",
		".jpg",
		".jpeg",
		".gif":
	default:
		return nil, fmt.Errorf(
			"formato %q não suportado; use PNG, JPG, JPEG ou GIF",
			extension,
		)
	}

	config, format, err :=
		image.DecodeConfig(
			bytes.NewReader(data),
		)

	if err != nil {
		return nil, fmt.Errorf(
			"imagem inválida: %w",
			err,
		)
	}

	frames := 1

	if format == "gif" {
		animation, err :=
			gif.DecodeAll(
				bytes.NewReader(data),
			)

		if err != nil {
			return nil, fmt.Errorf(
				"decodificar GIF: %w",
				err,
			)
		}

		frames =
			len(animation.Image)

		if frames < 1 {
			frames = 1
		}
	}

	if config.Width < 1 ||
		config.Height < 1 {
		return nil, fmt.Errorf(
			"dimensões inválidas: %dx%d",
			config.Width,
			config.Height,
		)
	}

	root, err := dataDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(
		root,
		0o700,
	); err != nil {
		return nil, fmt.Errorf(
			"criar pasta de imagens próprias: %w",
			err,
		)
	}

	destination :=
		filepath.Join(
			root,
			"source"+extension,
		)

	// Remove uma imagem própria anterior, independentemente
	// da extensão usada anteriormente.
	oldFiles, _ :=
		filepath.Glob(
			filepath.Join(
				root,
				"source.*",
			),
		)

	for _, old := range oldFiles {
		_ = os.Remove(old)
	}

	if err := os.WriteFile(
		destination,
		data,
		0o600,
	); err != nil {
		return nil, fmt.Errorf(
			"salvar imagem: %w",
			err,
		)
	}

	return &Info{
		Path: destination,

		Name: filepath.Base(
			originalName,
		),

		Format: strings.ToUpper(
			format,
		),

		Width: config.Width,

		Height: config.Height,

		Frames: frames,

		Size: int64(len(data)),
	}, nil
}

func dataDir() (string, error) {
	home, err :=
		os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf(
			"localizar HOME: %w",
			err,
		)
	}

	return filepath.Join(
		home,
		".local",
		"share",
		"minitela",
		"custom",
	), nil
}
