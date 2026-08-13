package device

import (
	"encoding/binary"
	"fmt"
)

// ReadFrame reads one complete AHMI/R15M protocol frame.
//
// controlFlag bits 0..14 contain:
//
//	len(commandType) + len(content)
//
// The frame also contains:
// start(2), control(2), CRC(2), footer(2).
func (p *SerialPort) ReadFrame() ([]byte, error) {
	header, err := p.ReadExact(4)
	if err != nil {
		return nil, err
	}

	if header[0] != 0x41 ||
		header[1] != 0x48 {
		return nil, fmt.Errorf(
			"cabeçalho serial inválido: %02X %02X",
			header[0],
			header[1],
		)
	}

	controlFlag := binary.BigEndian.Uint16(
		header[2:4],
	)

	commandLength := int(
		controlFlag & 0x7FFF,
	)

	if commandLength < 2 {
		return nil, fmt.Errorf(
			"commandLength inválido: %d",
			commandLength,
		)
	}

	// command/content + CRC + footer.
	remaining := commandLength + 4

	tail, err := p.ReadExact(remaining)
	if err != nil {
		return nil, err
	}

	frame := make(
		[]byte,
		0,
		len(header)+len(tail),
	)

	frame = append(frame, header...)
	frame = append(frame, tail...)

	return frame, nil
}
