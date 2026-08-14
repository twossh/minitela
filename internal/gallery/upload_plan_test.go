package gallery

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildUploadPlan(t *testing.T) {
	root := t.TempDir()

	texture :=
		filepath.Join(
			root,
			"Texture1.acf",
		)

	data := make(
		[]byte,
		2500,
	)

	// Footer ACF:
	// A5 5A 5A A5
	copy(
		data[len(data)-4:],
		[]byte{
			0xA5,
			0x5A,
			0x5A,
			0xA5,
		},
	)

	if err := os.WriteFile(
		texture,
		data,
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	item := Item{
		ID: 1,

		TexturePath: texture,

		TextureSize: int64(len(data)),
	}

	plan, err :=
		BuildUploadPlan(item)

	if err != nil {
		t.Fatal(err)
	}

	if plan.Address !=
		TextureAddress {
		t.Fatalf(
			"Address=0x%08X esperado=0x%08X",
			plan.Address,
			TextureAddress,
		)
	}

	if plan.FileSize != 2500 {
		t.Fatalf(
			"FileSize=%d esperado=2500",
			plan.FileSize,
		)
	}

	if plan.SimulatedChunks != 3 {
		t.Fatalf(
			"chunks=%d esperado=3",
			plan.SimulatedChunks,
		)
	}

	if plan.LastChunkSize != 452 {
		t.Fatalf(
			"LastChunkSize=%d esperado=452",
			plan.LastChunkSize,
		)
	}

	if fmt.Sprintf(
		"%x",
		plan.FileID,
	) == "" {
		t.Fatal(
			"MD5 vazio",
		)
	}
}
