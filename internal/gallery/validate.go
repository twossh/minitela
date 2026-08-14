package gallery

import (
	"fmt"
	"os"
)

const MaxTextureSize = 6436 * 1024

var acfFooter = []byte{
	0xA5,
	0x5A,
	0x5A,
	0xA5,
}

func ValidateTexture(
	item Item,
) error {
	info, err :=
		os.Stat(
			item.TexturePath,
		)

	if err != nil {
		return fmt.Errorf(
			"abrir textura: %w",
			err,
		)
	}

	if info.Size() <= 0 {
		return fmt.Errorf(
			"Texture%d.acf vazio",
			item.ID,
		)
	}

	if info.Size() >
		MaxTextureSize {
		return fmt.Errorf(
			"Texture%d.acf possui %.1f MB; máximo permitido %.1f MB",
			item.ID,
			float64(info.Size())/
				1024/
				1024,
			float64(MaxTextureSize)/
				1024/
				1024,
		)
	}

	file, err :=
		os.Open(
			item.TexturePath,
		)

	if err != nil {
		return err
	}

	defer file.Close()

	if info.Size() < 4 {
		return fmt.Errorf(
			"Texture%d.acf muito pequeno",
			item.ID,
		)
	}

	if _, err :=
		file.Seek(
			-4,
			2,
		); err != nil {
		return err
	}

	footer := make(
		[]byte,
		4,
	)

	if _, err :=
		file.Read(footer); err != nil {
		return err
	}

	for i := 0; i < 4; i++ {
		if footer[i] !=
			acfFooter[i] {
			return fmt.Errorf(
				"Texture%d.acf: footer ACF inválido: % X",
				item.ID,
				footer,
			)
		}
	}

	return nil
}
