package r15m

import (
	"fmt"

	"github.com/twossh/minitela/internal/device"
	"github.com/twossh/minitela/internal/protocol"
)

type Connection struct {
	Device            *device.Device
	Port              *device.SerialPort
	HandshakeResponse []byte
}

func (c *Connection) ReadRegister(
	regID uint16,
) (uint32, error) {
	if c == nil || c.Port == nil {
		return 0, fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	// Não existe ainda nenhum worker em segundo plano,
	// portanto podemos limpar qualquer byte antigo antes
	// de uma transação síncrona.
	_ = c.Port.ResetInputBuffer()

	request := protocol.BuildReadNumRegisterRequest(regID)

	if err := c.Port.WriteAll(request); err != nil {
		return 0, fmt.Errorf(
			"enviar leitura do registrador %d: %w",
			regID,
			err,
		)
	}

	response, err := c.Port.ReadExact(
		protocol.ReadRegisterResponseSize,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"ler registrador %d: %w",
			regID,
			err,
		)
	}

	responseRegID, value, err :=
		protocol.ParseReadNumRegisterResponse(response)

	if err != nil {
		return 0, fmt.Errorf(
			"decodificar registrador %d: %w",
			regID,
			err,
		)
	}

	if responseRegID != regID {
		return 0, fmt.Errorf(
			"registrador recebido=%d esperado=%d",
			responseRegID,
			regID,
		)
	}

	return value, nil
}

func Connect() (*Connection, error) {
	info, err := device.DetectR15M()
	if err != nil {
		return nil, fmt.Errorf(
			"detectar MiniTela: %w",
			err,
		)
	}

	port, err := device.OpenSerial(info.Path)
	if err != nil {
		return nil, err
	}

	success := false

	defer func() {
		if !success {
			_ = port.Close()
		}
	}()

	// Remove qualquer dado antigo que tenha ficado no buffer.
	_ = port.ResetInputBuffer()

	request := protocol.HandshakeRequest()

	if err := port.WriteAll(request); err != nil {
		return nil, fmt.Errorf(
			"enviar handshake: %w",
			err,
		)
	}

	response, err := port.ReadExact(
		protocol.HandshakeResponseSize,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"receber handshake: %w",
			err,
		)
	}

	if err := protocol.ValidateHandshakeResponse(response); err != nil {
		return nil, err
	}

	success = true

	return &Connection{
		Device:            info,
		Port:              port,
		HandshakeResponse: response,
	}, nil
}

func (c *Connection) Close() error {
	if c == nil || c.Port == nil {
		return nil
	}

	return c.Port.Close()
}
