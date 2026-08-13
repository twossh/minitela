package protocol

import (
	"encoding/binary"
	"fmt"
)

const ReadRegisterResponseSize = 17

// BuildReadNumRegisterRequest creates a request for one numeric register.
//
// R15M frame:
//
// 41 48          header
// 00             control
// 05             command length
// 00 90          GET_REGISTER
// C0             numeric read / one register
// RR RR          register ID, big-endian
// 00 00          CRC/reserved
// 4D 49          footer
func BuildReadNumRegisterRequest(regID uint16) []byte {
	frame := make([]byte, 13)

	frame[0] = 0x41
	frame[1] = 0x48

	frame[2] = 0x00
	frame[3] = 0x05

	frame[4] = 0x00
	frame[5] = 0x90

	frame[6] = 0xC0

	binary.BigEndian.PutUint16(frame[7:9], regID)

	frame[9] = 0x00
	frame[10] = 0x00

	frame[11] = 0x4D
	frame[12] = 0x49

	return frame
}

// ParseReadNumRegisterResponse parses one numeric-register response.
//
// Response content:
//
// byte 0       function/count header
// bytes 1..2   register ID
// bytes 3..6   uint32 value
func ParseReadNumRegisterResponse(
	frame []byte,
) (uint16, uint32, error) {
	if len(frame) != ReadRegisterResponseSize {
		return 0, 0, fmt.Errorf(
			"tamanho inválido: recebido=%d esperado=%d",
			len(frame),
			ReadRegisterResponseSize,
		)
	}

	if frame[0] != 0x41 || frame[1] != 0x48 {
		return 0, 0, fmt.Errorf(
			"cabeçalho inválido: %02X %02X",
			frame[0],
			frame[1],
		)
	}

	if frame[len(frame)-2] != 0x4D ||
		frame[len(frame)-1] != 0x49 {
		return 0, 0, fmt.Errorf(
			"rodapé inválido",
		)
	}

	// SET_REGISTER_RESPONSE / GET_REGISTER response.
	if frame[4] != 0x00 || frame[5] != 0xD0 {
		return 0, 0, fmt.Errorf(
			"comando de resposta inválido: %02X %02X",
			frame[4],
			frame[5],
		)
	}

	// Length 0x09 means:
	// command(2) + content(7).
	if frame[3] != 0x09 {
		return 0, 0, fmt.Errorf(
			"comprimento de comando inesperado: 0x%02X",
			frame[3],
		)
	}

	content := frame[6:13]

	header := content[0]

	functionCode := (header & 0x70) >> 4
	registerCount := int(header&0x0F) + 1

	if functionCode != 0 {
		return 0, 0, fmt.Errorf(
			"functionCode inesperado: %d",
			functionCode,
		)
	}

	if registerCount != 1 {
		return 0, 0, fmt.Errorf(
			"quantidade de registradores inesperada: %d",
			registerCount,
		)
	}

	regID := binary.BigEndian.Uint16(content[1:3])
	value := binary.BigEndian.Uint32(content[3:7])

	return regID, value, nil
}
