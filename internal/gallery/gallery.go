package gallery

import (
	"fmt"
	"os"
	"path/filepath"
)

const GallerySize = 21

type Item struct {
	ID int

	PreviewPath string
	TexturePath string

	PreviewSize int64
	TextureSize int64
}

func DefaultRoot() (string, error) {
	home, err := os.UserHomeDir()
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
		"vendor",
	), nil
}

func Load() ([]Item, error) {
	root, err := DefaultRoot()
	if err != nil {
		return nil, err
	}

	return LoadFrom(root)
}

func LoadFrom(root string) ([]Item, error) {
	items := make(
		[]Item,
		0,
		GallerySize,
	)

	for id := 1; id <= GallerySize; id++ {
		preview := filepath.Join(
			root,
			"previews",
			fmt.Sprintf("%d.gif", id),
		)

		texture := filepath.Join(
			root,
			"textures",
			fmt.Sprintf("Texture%d.acf", id),
		)

		previewInfo, err := os.Stat(preview)
		if err != nil {
			return nil, fmt.Errorf(
				"imagem %d: preview ausente: %w",
				id,
				err,
			)
		}

		if previewInfo.IsDir() {
			return nil, fmt.Errorf(
				"imagem %d: preview é diretório",
				id,
			)
		}

		textureInfo, err := os.Stat(texture)
		if err != nil {
			return nil, fmt.Errorf(
				"imagem %d: textura ausente: %w",
				id,
				err,
			)
		}

		if textureInfo.IsDir() {
			return nil, fmt.Errorf(
				"imagem %d: textura é diretório",
				id,
			)
		}

		if previewInfo.Size() == 0 {
			return nil, fmt.Errorf(
				"imagem %d: preview vazio",
				id,
			)
		}

		if textureInfo.Size() == 0 {
			return nil, fmt.Errorf(
				"imagem %d: textura vazia",
				id,
			)
		}

		items = append(
			items,
			Item{
				ID: id,

				PreviewPath: preview,

				TexturePath: texture,

				PreviewSize: previewInfo.Size(),

				TextureSize: textureInfo.Size(),
			},
		)
	}

	return items, nil
}

func Find(
	items []Item,
	id int,
) (Item, error) {
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}

	return Item{}, fmt.Errorf(
		"imagem %d não encontrada",
		id,
	)
}
