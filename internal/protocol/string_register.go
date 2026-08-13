package protocol

import (
	"encoding/binary"
	"fmt"
)

const maxStringRegisterSize = 1024

// BuildReadStringRegisterRequest creates a request to read
// a string register.
//
// Content:
//
// E0          string read
// RR RR       register ID
// LL LL       maximum requested length
func BuildReadStringRegisterRequest(
	regID uint16,
	maxLen uint16,
) []byte {
	const contentSize = 5

	commandLength := uint16(2 + contentSize)

	frame := make([]byte, 15)

	frame[0] = 0x41
	frame[1] = 0x48

	binary.BigEndian.PutUint16(
		frame[2:4],
		commandLength,
	)

	binary.BigEndian.PutUint16(
		frame[4:6],
		0x0090,
	)

	frame[6] = 0xE0

	binary.BigEndian.PutUint16(
		frame[7:9],
		regID,
	)

	binary.BigEndian.PutUint16(
		frame[9:11],
		maxLen,
	)

	// CRC disabled for this request.
	frame[11] = 0x00
	frame[12] = 0x00

	frame[13] = 0x4D
	frame[14] = 0x49

	return frame
}

// BuildWriteStringRegisterRequest creates a string SET_REGISTER
// request.
//
// Content:
//
// D0          string write
// RR RR       register ID
// LL LL       string length
// ...         string bytes
func BuildWriteStringRegisterRequest(
	regID uint16,
	data []byte,
) ([]byte, error) {
	if len(data) > maxStringRegisterSize {
		return nil, fmt.Errorf(
			"string muito grande: %d bytes; máximo=%d",
			len(data),
			maxStringRegisterSize,
		)
	}

	contentSize := 5 + len(data)
	commandLength := 2 + contentSize

	if commandLength > 0x7FFF {
		return nil, fmt.Errorf(
			"comando de string muito grande",
		)
	}

	frameSize :=
		2 + // start
			2 + // control
			commandLength +
			2 + // CRC
			2 // footer

	frame := make([]byte, frameSize)

	frame[0] = 0x41
	frame[1] = 0x48

	binary.BigEndian.PutUint16(
		frame[2:4],
		uint16(commandLength),
	)

	binary.BigEndian.PutUint16(
		frame[4:6],
		0x0090,
	)

	frame[6] = 0xD0

	binary.BigEndian.PutUint16(
		frame[7:9],
		regID,
	)

	binary.BigEndian.PutUint16(
		frame[9:11],
		uint16(len(data)),
	)

	copy(
		frame[11:11+len(data)],
		data,
	)

	crcOffset := 11 + len(data)

	// CRC disabled: field is still present.
	frame[crcOffset] = 0x00
	frame[crcOffset+1] = 0x00

	frame[crcOffset+2] = 0x4D
	frame[crcOffset+3] = 0x49

	return frame, nil
}

// ParseReadStringRegisterResponse parses a string register
// response returned by SET_REGISTER_RESPONSE.
func ParseReadStringRegisterResponse(
	frame []byte,
) (uint16, []byte, error) {
	if len(frame) < 14 {
		return 0, nil, fmt.Errorf(
			"resposta de string muito curta: %d",
			len(frame),
		)
	}

	if frame[0] != 0x41 ||
		frame[1] != 0x48 {
		return 0, nil, fmt.Errorf(
			"cabeçalho inválido",
		)
	}

	controlFlag := binary.BigEndian.Uint16(
		frame[2:4],
	)

	commandLength := int(
		controlFlag & 0x7FFF,
	)

	if commandLength < 7 {
		return 0, nil, fmt.Errorf(
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
		return 0, nil, fmt.Errorf(
			"tamanho inválido: recebido=%d esperado=%d",
			len(frame),
			expectedSize,
		)
	}

	commandType := binary.BigEndian.Uint16(
		frame[4:6],
	)

	if commandType != 0x00D0 {
		return 0, nil, fmt.Errorf(
			"resposta inesperada: 0x%04X",
			commandType,
		)
	}

	contentLength := commandLength - 2
	contentEnd := 6 + contentLength

	if frame[len(frame)-2] != 0x4D ||
		frame[len(frame)-1] != 0x49 {
		return 0, nil, fmt.Errorf(
			"rodapé inválido",
		)
	}

	// Validate CRC when the response actually contains one.
	if controlFlag&0x8000 != 0 {
		expectedCRC := binary.BigEndian.Uint16(
			frame[contentEnd : contentEnd+2],
		)

		// R15M SET_REGISTER_RESPONSE may advertise CRC
		// while returning 0000.
		if expectedCRC != 0 {
			actualCRC := CRC16IBM(
				frame[2:contentEnd],
			)

			if actualCRC != expectedCRC {
				return 0, nil, fmt.Errorf(
					"CRC inválido: recebido=0x%04X calculado=0x%04X",
					expectedCRC,
					actualCRC,
				)
			}
		}
	}

	content := frame[6:contentEnd]

	if len(content) < 5 {
		return 0, nil, fmt.Errorf(
			"conteúdo de string muito curto",
		)
	}

	regID := binary.BigEndian.Uint16(
		content[1:3],
	)

	strLen := int(
		binary.BigEndian.Uint16(
			content[3:5],
		),
	)

	if len(content) < 5+strLen {
		return 0, nil, fmt.Errorf(
			"string incompleta: recebido=%d esperado=%d",
			len(content)-5,
			strLen,
		)
	}

	result := make(
		[]byte,
		strLen,
	)

	copy(
		result,
		content[5:5+strLen],
	)

	return regID, result, nil
}
