package device

import (
	"fmt"
	"time"

	"go.bug.st/serial"
)

const (
	serialBaudRate = 115200
	readTimeout    = 2 * time.Second
)

type SerialPort struct {
	port serial.Port
	path string
}

func OpenSerial(path string) (*SerialPort, error) {
	mode := &serial.Mode{
		BaudRate: serialBaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(path, mode)
	if err != nil {
		return nil, fmt.Errorf(
			"abrir porta serial %s: %w",
			path,
			err,
		)
	}

	if err := port.SetReadTimeout(readTimeout); err != nil {
		_ = port.Close()

		return nil, fmt.Errorf(
			"configurar timeout serial: %w",
			err,
		)
	}

	return &SerialPort{
		port: port,
		path: path,
	}, nil
}

func (p *SerialPort) Path() string {
	return p.path
}

func (p *SerialPort) ResetInputBuffer() error {
	return p.port.ResetInputBuffer()
}

func (p *SerialPort) WriteAll(data []byte) error {
	written := 0

	for written < len(data) {
		n, err := p.port.Write(data[written:])
		if err != nil {
			return fmt.Errorf(
				"escrever na porta serial: %w",
				err,
			)
		}

		if n == 0 {
			return fmt.Errorf(
				"porta serial não aceitou dados",
			)
		}

		written += n
	}

	return nil
}

func (p *SerialPort) ReadExact(size int) ([]byte, error) {
	if size <= 0 {
		return nil, fmt.Errorf(
			"tamanho de leitura inválido: %d",
			size,
		)
	}

	data := make([]byte, size)
	total := 0

	for total < size {
		n, err := p.port.Read(data[total:])
		if err != nil {
			return nil, fmt.Errorf(
				"ler porta serial: %w",
				err,
			)
		}

		if n == 0 {
			return nil, fmt.Errorf(
				"timeout aguardando resposta: recebido=%d esperado=%d",
				total,
				size,
			)
		}

		total += n
	}

	return data, nil
}

func (p *SerialPort) Close() error {
	if p == nil || p.port == nil {
		return nil
	}

	return p.port.Close()
}
