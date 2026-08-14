package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	CommandGetDownloadStatus         uint16 = 0x0085
	CommandGetDownloadStatusResponse uint16 = 0x00C5

	CommandRequestDownload         uint16 = 0x0081
	CommandRequestDownloadResponse uint16 = 0x00C1

	CommandDownloadData         uint16 = 0x0082
	CommandDownloadDataResponse uint16 = 0x00C2

	CommandDownloadComplete         uint16 = 0x008F
	CommandDownloadCompleteResponse uint16 = 0x00CF

	CommandSwitchState         uint16 = 0x0071
	CommandSwitchStateResponse uint16 = 0x00B1
)

const (
	DownloadStatePreparing uint8 = 0x10
	DownloadStateActive    uint8 = 0x11
	DownloadStateAHMI      uint8 = 0x20

	ResponseProcessing uint32 = 0xFFFFFFFF
)

type DownloadStatus struct {
	Status uint8
	FileID [16]byte
	Offset uint32
}

type RequestDownloadResponse struct {
	MaxPageSize uint32
	Response    uint32
}

func buildUploadCommand(
	command uint16,
	content []byte,
) []byte {
	commandLength := 2 + len(content)

	frameSize :=
		2 +
			2 +
			commandLength +
			2 +
			2

	frame := make([]byte, frameSize)

	frame[0] = 0x41
	frame[1] = 0x48

	binary.BigEndian.PutUint16(
		frame[2:4],
		uint16(commandLength),
	)

	binary.BigEndian.PutUint16(
		frame[4:6],
		command,
	)

	copy(
		frame[6:6+len(content)],
		content,
	)

	// CRC desabilitado. O campo de dois bytes
	// continua presente no protocolo R15M.
	crcOffset := 6 + len(content)

	frame[crcOffset] = 0x00
	frame[crcOffset+1] = 0x00

	frame[crcOffset+2] = 0x4D
	frame[crcOffset+3] = 0x49

	return frame
}

func BuildGetDownloadStatusRequest() []byte {
	return buildUploadCommand(
		CommandGetDownloadStatus,
		nil,
	)
}

func BuildSwitchToDownloadModeRequest() []byte {
	return buildUploadCommand(
		CommandSwitchState,
		[]byte{0x10},
	)
}

func BuildRequestDownloadRequest(
	address uint32,
	fileSize uint32,
	fileID [16]byte,
) []byte {
	content := make([]byte, 24)

	binary.BigEndian.PutUint32(
		content[0:4],
		address,
	)

	binary.BigEndian.PutUint32(
		content[4:8],
		fileSize,
	)

	copy(
		content[8:24],
		fileID[:],
	)

	return buildUploadCommand(
		CommandRequestDownload,
		content,
	)
}

func BuildDownloadDataRequest(
	offset uint32,
	data []byte,
) []byte {
	padding :=
		(4 - (len(data) % 4)) % 4

	content := make(
		[]byte,
		4+len(data)+padding,
	)

	binary.BigEndian.PutUint32(
		content[0:4],
		offset,
	)

	copy(
		content[4:],
		data,
	)

	return buildUploadCommand(
		CommandDownloadData,
		content,
	)
}

func BuildDownloadCompleteRequest() []byte {
	return buildUploadCommand(
		CommandDownloadComplete,
		nil,
	)
}

func ParseUploadResponse(
	frame []byte,
	expectedCommand uint16,
) ([]byte, error) {
	if len(frame) < 10 {
		return nil, fmt.Errorf(
			"resposta muito curta: %d bytes",
			len(frame),
		)
	}

	if frame[0] != 0x41 ||
		frame[1] != 0x48 {
		return nil, fmt.Errorf(
			"cabeçalho inválido",
		)
	}

	control :=
		binary.BigEndian.Uint16(
			frame[2:4],
		)

	commandLength :=
		int(control & 0x7FFF)

	if commandLength < 2 {
		return nil, fmt.Errorf(
			"commandLength inválido: %d",
			commandLength,
		)
	}

	expectedSize :=
		2 +
			2 +
			commandLength +
			2 +
			2

	if len(frame) != expectedSize {
		return nil, fmt.Errorf(
			"tamanho inválido: recebido=%d esperado=%d",
			len(frame),
			expectedSize,
		)
	}

	command :=
		binary.BigEndian.Uint16(
			frame[4:6],
		)

	if command != expectedCommand {
		return nil, fmt.Errorf(
			"comando inesperado: recebido=0x%04X esperado=0x%04X",
			command,
			expectedCommand,
		)
	}

	if frame[len(frame)-2] != 0x4D ||
		frame[len(frame)-1] != 0x49 {
		return nil, fmt.Errorf(
			"rodapé inválido",
		)
	}

	contentLength :=
		commandLength - 2

	contentEnd :=
		6 + contentLength

	if control&0x8000 != 0 {
		expectedCRC :=
			binary.BigEndian.Uint16(
				frame[contentEnd : contentEnd+2],
			)

		// Alguns ACKs do R15M anunciam CRC,
		// porém retornam 0000.
		if expectedCRC != 0 {
			actualCRC :=
				CRC16IBM(
					frame[2:contentEnd],
				)

			if expectedCRC != actualCRC {
				return nil, fmt.Errorf(
					"CRC inválido: recebido=0x%04X calculado=0x%04X",
					expectedCRC,
					actualCRC,
				)
			}
		}
	}

	result := make(
		[]byte,
		contentLength,
	)

	copy(
		result,
		frame[6:contentEnd],
	)

	return result, nil
}

func ParseDownloadStatus(
	content []byte,
) (*DownloadStatus, error) {
	if len(content) < 21 {
		return nil, fmt.Errorf(
			"DownloadStatus curto: %d bytes",
			len(content),
		)
	}

	result := &DownloadStatus{
		Status: content[0],

		Offset: binary.BigEndian.Uint32(
			content[17:21],
		),
	}

	copy(
		result.FileID[:],
		content[1:17],
	)

	return result, nil
}

func ParseRequestDownloadResponse(
	content []byte,
) (*RequestDownloadResponse, error) {
	if len(content) < 8 {
		return nil, fmt.Errorf(
			"RequestDownloadResponse curto: %d bytes",
			len(content),
		)
	}

	return &RequestDownloadResponse{
		MaxPageSize: binary.BigEndian.Uint32(
			content[0:4],
		),

		Response: binary.BigEndian.Uint32(
			content[4:8],
		),
	}, nil
}

func ParseUploadResultCode(
	content []byte,
) (uint32, error) {
	if len(content) < 4 {
		return 0, fmt.Errorf(
			"resposta sem código: %d bytes",
			len(content),
		)
	}

	return binary.BigEndian.Uint32(
		content[0:4],
	), nil
}
