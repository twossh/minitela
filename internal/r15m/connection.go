package r15m

import (
	"fmt"

	"github.com/twossh/minitela/internal/device"
	"github.com/twossh/minitela/internal/protocol"
)

// Connection represents an active connection with the
// Positivo R15M auxiliary display.
type Connection struct {
	Device            *device.Device
	Port              *device.SerialPort
	HandshakeResponse []byte
}

// Connect detects the Positivo R15M, opens its serial port
// and performs the initial handshake.
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

	// Remove any stale data that may be present in
	// the serial input buffer.
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
		return nil, fmt.Errorf(
			"validar handshake: %w",
			err,
		)
	}

	success = true

	return &Connection{
		Device:            info,
		Port:              port,
		HandshakeResponse: response,
	}, nil
}

// ReadRegister reads one numeric register from the R15M.
func (c *Connection) ReadRegister(
	regID uint16,
) (uint32, error) {
	if c == nil || c.Port == nil {
		return 0, fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	// Currently all operations are synchronous.
	// Clear stale bytes before starting a new transaction.
	_ = c.Port.ResetInputBuffer()

	request := protocol.BuildReadNumRegisterRequest(
		regID,
	)

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
		protocol.ParseReadNumRegisterResponse(
			response,
		)
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

// WriteRegister writes one numeric register to the R15M.
//
// The R15M returns a SET_REGISTER_RESPONSE ACK.
// The ACK does not necessarily contain the written value,
// therefore it is only validated here.
func (c *Connection) WriteRegister(
	regID uint16,
	value uint32,
) error {
	if c == nil || c.Port == nil {
		return fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	// Remove stale bytes before starting the transaction.
	_ = c.Port.ResetInputBuffer()

	request :=
		protocol.BuildWriteNumRegisterRequest(
			regID,
			value,
		)

	if err := c.Port.WriteAll(request); err != nil {
		return fmt.Errorf(
			"escrever registrador %d: %w",
			regID,
			err,
		)
	}

	// SET_REGISTER responses may have variable frame sizes,
	// therefore ReadFrame() is used instead of ReadExact().
	response, err := c.Port.ReadFrame()
	if err != nil {
		return fmt.Errorf(
			"receber ACK do registrador %d: %w",
			regID,
			err,
		)
	}

	if err :=
		protocol.ValidateWriteNumRegisterResponse(
			response,
		); err != nil {
		return fmt.Errorf(
			"ACK do registrador %d: %w",
			regID,
			err,
		)
	}

	return nil
}

// WriteRegisterVerified writes a numeric register and then reads
// it back to confirm that the requested value was actually applied.
func (c *Connection) WriteRegisterVerified(
	regID uint16,
	value uint32,
) (uint32, error) {
	if err := c.WriteRegister(
		regID,
		value,
	); err != nil {
		return 0, err
	}

	actual, err := c.ReadRegister(
		regID,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"verificar registrador %d: %w",
			regID,
			err,
		)
	}

	if actual != value {
		return actual, fmt.Errorf(
			"verificação falhou: registrador=%d escrito=%d lido=%d",
			regID,
			value,
			actual,
		)
	}

	return actual, nil
}

// Close closes the serial connection.
func (c *Connection) Close() error {
	if c == nil || c.Port == nil {
		return nil
	}

	return c.Port.Close()
}
