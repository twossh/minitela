package r15m

import (
	"bytes"
	"fmt"

	"github.com/twossh/minitela/internal/protocol"
)

// ReadStringRegister reads a textual register from the R15M.
//
// The firmware may return a fixed-size buffer whose length corresponds
// to the requested maximum size rather than the useful string length.
// Display strings are NUL-terminated, therefore everything after the
// first 0x00 byte must be ignored.
func (c *Connection) ReadStringRegister(
	regID uint16,
	maxLen uint16,
) ([]byte, error) {
	raw, err := c.ReadRawStringRegister(
		regID,
		maxLen,
	)
	if err != nil {
		return nil, err
	}

	return cleanDisplayString(raw), nil
}

// ReadRawStringRegister returns the complete string buffer exactly as
// supplied by the R15M firmware. This is useful for diagnostics and
// future reverse engineering.
func (c *Connection) ReadRawStringRegister(
	regID uint16,
	maxLen uint16,
) ([]byte, error) {
	if c == nil || c.Port == nil {
		return nil, fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	_ = c.Port.ResetInputBuffer()

	request :=
		protocol.BuildReadStringRegisterRequest(
			regID,
			maxLen,
		)

	if err := c.Port.WriteAll(request); err != nil {
		return nil, fmt.Errorf(
			"enviar leitura da string %d: %w",
			regID,
			err,
		)
	}

	response, err := c.Port.ReadFrame()
	if err != nil {
		return nil, fmt.Errorf(
			"ler string %d: %w",
			regID,
			err,
		)
	}

	responseRegID, value, err :=
		protocol.ParseReadStringRegisterResponse(
			response,
		)
	if err != nil {
		return nil, fmt.Errorf(
			"decodificar string %d: %w",
			regID,
			err,
		)
	}

	if responseRegID != regID {
		return nil, fmt.Errorf(
			"registrador recebido=%d esperado=%d",
			responseRegID,
			regID,
		)
	}

	return value, nil
}

func (c *Connection) WriteStringRegister(
	regID uint16,
	value []byte,
) error {
	if c == nil || c.Port == nil {
		return fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	request, err :=
		protocol.BuildWriteStringRegisterRequest(
			regID,
			value,
		)
	if err != nil {
		return err
	}

	_ = c.Port.ResetInputBuffer()

	if err := c.Port.WriteAll(request); err != nil {
		return fmt.Errorf(
			"escrever string %d: %w",
			regID,
			err,
		)
	}

	response, err := c.Port.ReadFrame()
	if err != nil {
		return fmt.Errorf(
			"receber ACK da string %d: %w",
			regID,
			err,
		)
	}

	if err :=
		protocol.ValidateWriteNumRegisterResponse(
			response,
		); err != nil {
		return fmt.Errorf(
			"ACK da string %d: %w",
			regID,
			err,
		)
	}

	return nil
}

// cleanDisplayString converts the fixed-size string buffer used by
// the R15M into the actual textual value displayed on screen.
func cleanDisplayString(data []byte) []byte {
	if index := bytes.IndexByte(data, 0x00); index >= 0 {
		data = data[:index]
	}

	result := make([]byte, len(data))
	copy(result, data)

	return result
}
