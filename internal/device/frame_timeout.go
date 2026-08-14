package device

import (
	"encoding/binary"
	"fmt"
	"time"
)

// ReadFrameTimeout lê um frame completo do protocolo R15M,
// permitindo um timeout total diferente do timeout serial padrão.
//
// O ReadFrame() normal continua inalterado e continua sendo usado
// pelas operações rápidas como registradores, Monitor e Clima.
func (p *SerialPort) ReadFrameTimeout(
	timeout time.Duration,
) ([]byte, error) {
	if p == nil || p.port == nil {
		return nil, fmt.Errorf(
			"porta serial não inicializada",
		)
	}

	if timeout <= 0 {
		return nil, fmt.Errorf(
			"timeout inválido: %s",
			timeout,
		)
	}

	// Usamos leituras curtas internamente para podermos
	// respeitar o timeout total desta operação.
	const pollTimeout = 250 * time.Millisecond

	if err := p.port.SetReadTimeout(
		pollTimeout,
	); err != nil {
		return nil, fmt.Errorf(
			"configurar timeout temporário: %w",
			err,
		)
	}

	// Sempre restaura os 2 segundos usados pelo restante
	// do MiniTela.
	defer func() {
		_ = p.port.SetReadTimeout(
			readTimeout,
		)
	}()

	deadline := time.Now().Add(
		timeout,
	)

	header, err := p.readExactUntil(
		4,
		deadline,
	)
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

	controlFlag :=
		binary.BigEndian.Uint16(
			header[2:4],
		)

	commandLength :=
		int(
			controlFlag & 0x7FFF,
		)

	if commandLength < 2 {
		return nil, fmt.Errorf(
			"commandLength inválido: %d",
			commandLength,
		)
	}

	// command/content + CRC + footer.
	remaining :=
		commandLength + 4

	tail, err := p.readExactUntil(
		remaining,
		deadline,
	)
	if err != nil {
		return nil, err
	}

	frame := make(
		[]byte,
		0,
		len(header)+len(tail),
	)

	frame = append(
		frame,
		header...,
	)

	frame = append(
		frame,
		tail...,
	)

	return frame, nil
}

func (p *SerialPort) readExactUntil(
	size int,
	deadline time.Time,
) ([]byte, error) {
	if size <= 0 {
		return []byte{}, nil
	}

	data := make(
		[]byte,
		size,
	)

	total := 0

	for total < size {
		if time.Now().After(
			deadline,
		) {
			return nil, fmt.Errorf(
				"timeout aguardando resposta: recebido=%d esperado=%d",
				total,
				size,
			)
		}

		n, err := p.port.Read(
			data[total:],
		)

		if err != nil {
			return nil, fmt.Errorf(
				"ler porta serial: %w",
				err,
			)
		}

		if n == 0 {
			continue
		}

		total += n
	}

	return data, nil
}
