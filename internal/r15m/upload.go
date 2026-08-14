package r15m

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/twossh/minitela/internal/protocol"
)

const TextureAddress uint32 = 0x08100000

type UploadProgress struct {
	BytesSent  int64
	TotalBytes int64
	Percent    int
}

type UploadResult struct {
	FileID [16]byte

	FileSize uint32

	Address uint32

	MaxPageSize uint32

	StartOffset uint32

	AlreadyPresent bool
}

func (c *Connection) UploadTexture(
	ctx context.Context,
	data []byte,
	onProgress func(UploadProgress),
) (*UploadResult, error) {
	if c == nil || c.Port == nil {
		return nil, fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf(
			"arquivo ACF vazio",
		)
	}

	const maxSize = 6436 * 1024

	if len(data) > maxSize {
		return nil, fmt.Errorf(
			"arquivo muito grande: %d bytes; máximo=%d",
			len(data),
			maxSize,
		)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fileID := md5.Sum(data)
	fileSize := uint32(len(data))

	result := &UploadResult{
		FileID:   fileID,
		FileSize: fileSize,
		Address:  TextureAddress,
	}

	// O Connect() já realizou o handshake.
	status, err :=
		c.ProbeDownloadStatus()

	if err != nil {
		return nil, fmt.Errorf(
			"GetDownloadStatus: %w",
			err,
		)
	}

	startOffset := uint32(0)

	switch status.Status {
	case protocol.DownloadStatePreparing,
		protocol.DownloadStateActive:

		if status.FileID == fileID {
			if status.Offset <= fileSize {
				startOffset =
					status.Offset
			}
		}

	case protocol.DownloadStateAHMI:
		if status.FileID == fileID {
			result.AlreadyPresent = true
			result.StartOffset =
				fileSize

			reportUploadProgress(
				onProgress,
				int64(fileSize),
				int64(fileSize),
			)

			return result, nil
		}

		if err :=
			c.switchToDownloadMode(
				ctx,
			); err != nil {
			return nil, fmt.Errorf(
				"SwitchState: %w",
				err,
			)
		}

	default:
		return nil, fmt.Errorf(
			"estado de download desconhecido: 0x%02X",
			status.Status,
		)
	}

	result.StartOffset =
		startOffset

	requestResult, err :=
		c.requestTextureDownload(
			ctx,
			TextureAddress,
			fileSize,
			fileID,
		)

	if err != nil {
		return nil, err
	}

	if requestResult.MaxPageSize == 0 {
		return nil, fmt.Errorf(
			"R15M retornou MaxPageSize=0",
		)
	}

	result.MaxPageSize =
		requestResult.MaxPageSize

	if requestResult.Response ==
		protocol.ResponseProcessing {

		if err :=
			c.waitRequestDownloadReady(
				ctx,
			); err != nil {
			return nil, err
		}

	} else if requestResult.Response != 0 {
		return nil, fmt.Errorf(
			"RequestDownload rejeitado: 0x%08X",
			requestResult.Response,
		)
	}

	offset := startOffset

	lastPercent := -1

	for offset < fileSize {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		end :=
			offset +
				requestResult.MaxPageSize

		if end > fileSize {
			end = fileSize
		}

		chunk :=
			data[offset:end]

		if err :=
			c.sendTextureChunk(
				ctx,
				offset,
				chunk,
			); err != nil {
			return nil, fmt.Errorf(
				"DownloadData offset=%d: %w",
				offset,
				err,
			)
		}

		offset = end

		percent :=
			int(
				100 *
					uint64(offset) /
					uint64(fileSize),
			)

		// Atualiza a UI/CLI de 5 em 5%.
		if percent/5 >
			lastPercent/5 {

			lastPercent =
				percent

			reportUploadProgress(
				onProgress,
				int64(offset),
				int64(fileSize),
			)
		}
	}

	reportUploadProgress(
		onProgress,
		int64(fileSize),
		int64(fileSize),
	)

	// O SideCar original deliberadamente não torna
	// uma falha de DownloadComplete fatal depois que
	// todos os bytes foram transferidos.
	//
	// Alguns firmwares reiniciam/desconectam a USB
	// imediatamente nessa etapa.
	_ = c.downloadComplete(ctx)

	return result, nil
}

func (c *Connection) switchToDownloadMode(
	ctx context.Context,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_ = c.Port.ResetInputBuffer()

	if err :=
		c.Port.WriteAll(
			protocol.
				BuildSwitchToDownloadModeRequest(),
		); err != nil {
		return err
	}

	frame, err :=
		c.Port.ReadFrameTimeout(3 * time.Second)

	if err != nil {
		return err
	}

	content, err :=
		protocol.ParseUploadResponse(
			frame,
			protocol.
				CommandSwitchStateResponse,
		)

	if err != nil {
		return err
	}

	code, err :=
		protocol.ParseUploadResultCode(
			content,
		)

	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf(
			"código=0x%08X",
			code,
		)
	}

	return nil
}

func (c *Connection) requestTextureDownload(
	ctx context.Context,
	address uint32,
	fileSize uint32,
	fileID [16]byte,
) (*protocol.RequestDownloadResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	_ = c.Port.ResetInputBuffer()

	request :=
		protocol.BuildRequestDownloadRequest(
			address,
			fileSize,
			fileID,
		)

	if err :=
		c.Port.WriteAll(request); err != nil {
		return nil, fmt.Errorf(
			"enviar RequestDownload: %w",
			err,
		)
	}

	frame, err :=
		c.Port.ReadFrameTimeout(60 * time.Second)

	if err != nil {
		return nil, fmt.Errorf(
			"ler RequestDownloadResponse: %w",
			err,
		)
	}

	content, err :=
		protocol.ParseUploadResponse(
			frame,
			protocol.
				CommandRequestDownloadResponse,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"validar RequestDownloadResponse: %w",
			err,
		)
	}

	result, err :=
		protocol.ParseRequestDownloadResponse(
			content,
		)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Connection) waitRequestDownloadReady(
	ctx context.Context,
) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		frame, err :=
			c.Port.ReadFrameTimeout(60 * time.Second)

		if err != nil {
			return fmt.Errorf(
				"aguardar RequestDownload: %w",
				err,
			)
		}

		content, err :=
			protocol.ParseUploadResponse(
				frame,
				protocol.
					CommandRequestDownloadResponse,
			)

		if err != nil {
			return err
		}

		result, err :=
			protocol.ParseRequestDownloadResponse(
				content,
			)

		if err != nil {
			return err
		}

		switch result.Response {
		case 0:
			return nil

		case protocol.ResponseProcessing:
			continue

		default:
			return fmt.Errorf(
				"RequestDownload processamento: 0x%08X",
				result.Response,
			)
		}
	}
}

func (c *Connection) sendTextureChunk(
	ctx context.Context,
	offset uint32,
	chunk []byte,
) error {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {

		if err := ctx.Err(); err != nil {
			return err
		}

		_ = c.Port.ResetInputBuffer()

		request :=
			protocol.BuildDownloadDataRequest(
				offset,
				chunk,
			)

		if err :=
			c.Port.WriteAll(request); err != nil {
			lastErr = err
			continue
		}

		frame, err :=
			c.Port.ReadFrameTimeout(5 * time.Second)

		if err != nil {
			lastErr = err
			continue
		}

		content, err :=
			protocol.ParseUploadResponse(
				frame,
				protocol.
					CommandDownloadDataResponse,
			)

		if err != nil {
			lastErr = err
			continue
		}

		code, err :=
			protocol.ParseUploadResultCode(
				content,
			)

		if err != nil {
			lastErr = err
			continue
		}

		switch code {
		case 0:
			return nil

		case protocol.ResponseProcessing:
			if err :=
				c.waitChunkReady(
					ctx,
				); err != nil {

				lastErr = err
				continue
			}

			return nil

		default:
			lastErr =
				fmt.Errorf(
					"resposta=0x%08X",
					code,
				)
		}
	}

	return fmt.Errorf(
		"3 tentativas falharam: %w",
		lastErr,
	)
}

func (c *Connection) waitChunkReady(
	ctx context.Context,
) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		frame, err :=
			c.Port.ReadFrameTimeout(5 * time.Second)

		if err != nil {
			return err
		}

		content, err :=
			protocol.ParseUploadResponse(
				frame,
				protocol.
					CommandDownloadDataResponse,
			)

		if err != nil {
			return err
		}

		code, err :=
			protocol.ParseUploadResultCode(
				content,
			)

		if err != nil {
			return err
		}

		switch code {
		case 0:
			return nil

		case protocol.ResponseProcessing:
			continue

		default:
			return fmt.Errorf(
				"DownloadData processamento: 0x%08X",
				code,
			)
		}
	}
}

func (c *Connection) downloadComplete(
	ctx context.Context,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_ = c.Port.ResetInputBuffer()

	if err :=
		c.Port.WriteAll(
			protocol.
				BuildDownloadCompleteRequest(),
		); err != nil {
		return err
	}

	frame, err :=
		c.Port.ReadFrameTimeout(30 * time.Second)

	if err != nil {
		return err
	}

	content, err :=
		protocol.ParseUploadResponse(
			frame,
			protocol.
				CommandDownloadCompleteResponse,
		)

	if err != nil {
		return err
	}

	code, err :=
		protocol.ParseUploadResultCode(
			content,
		)

	if err != nil {
		return err
	}

	for code ==
		protocol.ResponseProcessing {

		if err := ctx.Err(); err != nil {
			return err
		}

		frame, err =
			c.Port.ReadFrameTimeout(30 * time.Second)

		if err != nil {
			return err
		}

		content, err =
			protocol.ParseUploadResponse(
				frame,
				protocol.
					CommandDownloadCompleteResponse,
			)

		if err != nil {
			return err
		}

		code, err =
			protocol.ParseUploadResultCode(
				content,
			)

		if err != nil {
			return err
		}
	}

	if code != 0 {
		return fmt.Errorf(
			"DownloadComplete código=0x%08X",
			code,
		)
	}

	return nil
}

func reportUploadProgress(
	fn func(UploadProgress),
	sent int64,
	total int64,
) {
	if fn == nil {
		return
	}

	percent := 100

	if total > 0 {
		percent =
			int(
				100 *
					sent /
					total,
			)
	}

	fn(
		UploadProgress{
			BytesSent: sent,

			TotalBytes: total,

			Percent: percent,
		},
	)
}
