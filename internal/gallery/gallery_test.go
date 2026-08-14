package gallery

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFrom(t *testing.T) {
	root := t.TempDir()

	previews := filepath.Join(
		root,
		"previews",
	)

	textures := filepath.Join(
		root,
		"textures",
	)

	if err := os.MkdirAll(
		previews,
		0o755,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(
		textures,
		0o755,
	); err != nil {
		t.Fatal(err)
	}

	for id := 1; id <= GallerySize; id++ {
		preview := filepath.Join(
			previews,
			fmt.Sprintf("%d.gif", id),
		)

		texture := filepath.Join(
			textures,
			fmt.Sprintf("Texture%d.acf", id),
		)

		if err := os.WriteFile(
			preview,
			[]byte("GIF"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(
			texture,
			[]byte("ACF"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
	}

	items, err := LoadFrom(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != GallerySize {
		t.Fatalf(
			"itens=%d esperado=%d",
			len(items),
			GallerySize,
		)
	}

	if items[0].ID != 1 {
		t.Fatalf(
			"primeiro ID=%d esperado=1",
			items[0].ID,
		)
	}

	if items[20].ID != 21 {
		t.Fatalf(
			"último ID=%d esperado=21",
			items[20].ID,
		)
	}
}

func TestFind(t *testing.T) {
	items := []Item{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	item, err := Find(items, 2)
	if err != nil {
		t.Fatal(err)
	}

	if item.ID != 2 {
		t.Fatalf(
			"ID=%d esperado=2",
			item.ID,
		)
	}
}
