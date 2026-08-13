package protocol

import (
	"bytes"
	"fmt"
)

const HandshakeResponseSize = 14

var handshakeRequest = [...]byte{
	0x41, 0x48,
	0x00, 0x02,
	0x00, 0x80,
	0x00, 0x00,
	0x4D, 0x49,
}

var handshakeResponse = [...]byte{
	0x41, 0x48,
	0x00, 0x06,
	0x00, 0xC0,
	0x00, 0x00,
	0x04, 0x00, 0x00, 0x00,
	0x4D, 0x49,
}

func HandshakeRequest() []byte {
	data := make([]byte, len(handshakeRequest))
	copy(data, handshakeRequest[:])
	return data
}

func ValidateHandshakeResponse(data []byte) error {
	if len(data) != len(handshakeResponse) {
		return fmt.Errorf(
			"tamanho inválido da resposta: recebido=%d esperado=%d",
			len(data),
			len(handshakeResponse),
		)
	}

	if !bytes.Equal(data, handshakeResponse[:]) {
		return fmt.Errorf(
			"resposta de handshake inválida: % X",
			data,
		)
	}

	return nil
}
