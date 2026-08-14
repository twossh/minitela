package r15m

import (
	"fmt"

	"github.com/twossh/minitela/internal/protocol"
)

// ProbeDownloadStatus consulta somente o estado atual
// do mecanismo de download do R15M.
//
// Esta função NÃO inicia upload e NÃO altera memória.
func (c *Connection) ProbeDownloadStatus() (
	*protocol.DownloadStatus,
	error,
) {
	if c == nil || c.Port == nil {
		return nil, fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	_ = c.Port.ResetInputBuffer()

	request :=
		protocol.BuildGetDownloadStatusRequest()

	if err := c.Port.WriteAll(
		request,
	); err != nil {
		return nil, fmt.Errorf(
			"enviar GetDownloadStatus: %w",
			err,
		)
	}

	frame, err :=
		c.Port.ReadFrame()

	if err != nil {
		return nil, fmt.Errorf(
			"ler GetDownloadStatusResponse: %w",
			err,
		)
	}

	content, err :=
		protocol.ParseUploadResponse(
			frame,
			protocol.CommandGetDownloadStatusResponse,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"validar GetDownloadStatusResponse: %w",
			err,
		)
	}

	status, err :=
		protocol.ParseDownloadStatus(
			content,
		)

	if err != nil {
		return nil, fmt.Errorf(
			"interpretar GetDownloadStatus: %w",
			err,
		)
	}

	return status, nil
}
