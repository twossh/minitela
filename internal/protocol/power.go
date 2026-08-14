package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	CommandReboot         uint16 = 0x0070
	CommandRebootResponse uint16 = 0x00B0
)

// BuildRebootRequest cria:
//
// 41 48 00 02 00 70 00 00 4D 49
func BuildRebootRequest() []byte {
	return []byte{
		0x41, 0x48,
		0x00, 0x02,
		0x00, 0x70,
		0x00, 0x00,
		0x4D, 0x49,
	}
}

func ValidateRebootResponse(
	frame []byte,
) error {
	if len(frame) < 10 {
		return fmt.Errorf(
			"resposta de reboot muito curta: %d bytes",
			len(frame),
		)
	}

	if frame[0] != 0x41 ||
		frame[1] != 0x48 {
		return fmt.Errorf(
			"cabeçalho de reboot inválido",
		)
	}

	control :=
		binary.BigEndian.Uint16(
			frame[2:4],
		)

	contentLen :=
		int(control & 0x7FFF)

	expectedSize :=
		2 + 2 + contentLen + 2 + 2

	if len(frame) != expectedSize {
		return fmt.Errorf(
			"tamanho de reboot inválido: recebido=%d esperado=%d",
			len(frame),
			expectedSize,
		)
	}

	if contentLen < 2 {
		return fmt.Errorf(
			"resposta de reboot sem command type",
		)
	}

	command :=
		binary.BigEndian.Uint16(
			frame[4:6],
		)

	if command != CommandRebootResponse {
		return fmt.Errorf(
			"resposta inesperada ao reboot: 0x%04X",
			command,
		)
	}

	if frame[len(frame)-2] != 0x4D ||
		frame[len(frame)-1] != 0x49 {
		return fmt.Errorf(
			"rodapé de reboot inválido",
		)
	}

	return nil
}
