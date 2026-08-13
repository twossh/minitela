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

// BuildWriteNumRegisterRequest creates a CRC-enabled numeric SET_REGISTER
// command for one R15M register.
//
// Frame:
//
// 41 48          start
// 80 09          CRC enabled + command/content length
// 00 90          SET_REGISTER
// 80             numeric write, one register
// RR RR          register ID
// VV VV VV VV    uint32 value
// CC CC          CRC-16/IBM
// 4D 49          end
func BuildWriteNumRegisterRequest(
	regID uint16,
	value uint32,
) []byte {
	const (
		commandLength = 9
		frameSize     = 17
	)

	frame := make([]byte, frameSize)

	frame[0] = 0x41
	frame[1] = 0x48

	// Bit 15 = CRC enabled.
	controlFlag := uint16(0x8000 | commandLength)

	binary.BigEndian.PutUint16(
		frame[2:4],
		controlFlag,
	)

	// SET_REGISTER.
	binary.BigEndian.PutUint16(
		frame[4:6],
		0x0090,
	)

	// Numeric write / one register.
	frame[6] = 0x80

	binary.BigEndian.PutUint16(
		frame[7:9],
		regID,
	)

	binary.BigEndian.PutUint32(
		frame[9:13],
		value,
	)

	// CRC covers:
	// controlFlag + command type + content.
	crc := CRC16IBM(frame[2:13])

	binary.BigEndian.PutUint16(
		frame[13:15],
		crc,
	)

	frame[15] = 0x4D
	frame[16] = 0x49

	return frame
}

// ValidateWriteNumRegisterResponse validates SET_REGISTER_RESPONSE.
//
// R15M quirk:
// the firmware may set the CRC flag but return CRC 0x0000.
// A valid SET_REGISTER_RESPONSE is treated as an ACK.
func ValidateWriteNumRegisterResponse(
	frame []byte,
) error {
	if len(frame) < 10 {
		return fmt.Errorf(
			"resposta SET_REGISTER muito curta: %d bytes",
			len(frame),
		)
	}

	if frame[0] != 0x41 || frame[1] != 0x48 {
		return fmt.Errorf(
			"cabeçalho SET_REGISTER inválido",
		)
	}

	controlFlag := binary.BigEndian.Uint16(frame[2:4])

	crcEnabled := controlFlag&0x8000 != 0
	commandLength := int(controlFlag & 0x7FFF)

	if commandLength < 2 {
		return fmt.Errorf(
			"commandLength inválido: %d",
			commandLength,
		)
	}

	expectedSize :=
		2 + // start
			2 + // controlFlag
			commandLength +
			2 + // CRC
			2 // footer

	if len(frame) != expectedSize {
		return fmt.Errorf(
			"tamanho SET_REGISTER inválido: recebido=%d esperado=%d",
			len(frame),
			expectedSize,
		)
	}

	if frame[len(frame)-2] != 0x4D ||
		frame[len(frame)-1] != 0x49 {
		return fmt.Errorf(
			"rodapé SET_REGISTER inválido",
		)
	}

	commandType := binary.BigEndian.Uint16(
		frame[4:6],
	)

	if commandType != 0x00D0 {
		return fmt.Errorf(
			"resposta inesperada: comando=0x%04X esperado=0x00D0",
			commandType,
		)
	}

	crcOffset := 4 + commandLength

	expectedCRC := binary.BigEndian.Uint16(
		frame[crcOffset : crcOffset+2],
	)

	if crcEnabled && expectedCRC != 0x0000 {
		actualCRC := CRC16IBM(
			frame[2:crcOffset],
		)

		if actualCRC != expectedCRC {
			return fmt.Errorf(
				"CRC inválido: recebido=0x%04X calculado=0x%04X",
				expectedCRC,
				actualCRC,
			)
		}
	}

	return nil
}
