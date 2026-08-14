package gallery

import (
	"crypto/md5"
	"fmt"
	"os"
)

const (
	TextureAddress uint32 = 0x08100000

	// Valor apenas para simulação.
	// O tamanho real de página será informado pelo R15M
	// durante RequestDownload no upload físico.
	DryRunPageSize = 1024
)

type UploadPlan struct {
	Item Item

	Address uint32

	FileSize int64

	FileID [16]byte

	SimulatedPageSize int

	SimulatedChunks int

	LastChunkSize int
}

func BuildUploadPlan(
	item Item,
) (*UploadPlan, error) {
	if err := ValidateTexture(item); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(
		item.TexturePath,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"ler Texture%d.acf: %w",
			item.ID,
			err,
		)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf(
			"Texture%d.acf vazio",
			item.ID,
		)
	}

	fileID := md5.Sum(data)

	chunks :=
		len(data) / DryRunPageSize

	if len(data)%DryRunPageSize != 0 {
		chunks++
	}

	lastChunk :=
		len(data) % DryRunPageSize

	if lastChunk == 0 {
		lastChunk = DryRunPageSize
	}

	return &UploadPlan{
		Item: item,

		Address: TextureAddress,

		FileSize: int64(len(data)),

		FileID: fileID,

		SimulatedPageSize: DryRunPageSize,

		SimulatedChunks: chunks,

		LastChunkSize: lastChunk,
	}, nil
}
