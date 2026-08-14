package r15m

import (
	"fmt"

	"github.com/twossh/minitela/internal/protocol"
)

func (c *Connection) PowerOff() error {
	if c == nil || c.Port == nil {
		return fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	_, err := c.WriteRegisterVerified(
		RegisterBrightness,
		0,
	)

	if err != nil {
		return fmt.Errorf(
			"desligar MiniTela: %w",
			err,
		)
	}

	return nil
}

// Reboot reinicia o controlador da MiniTela.
//
// O R15M normalmente derruba a interface USB imediatamente
// depois de receber o comando 0x0070.
//
// Por isso NÃO devemos exigir a leitura do CmdRebootResponse.
// Em alguns firmwares o ACK pode existir, mas a porta USB pode
// desaparecer antes que o sistema operacional consiga recebê-lo.
//
// Se WriteAll() terminou com sucesso, consideramos o comando
// entregue ao controlador.
func (c *Connection) Reboot() error {
	if c == nil || c.Port == nil {
		return fmt.Errorf(
			"MiniTela não conectada",
		)
	}

	_ = c.Port.ResetInputBuffer()

	request :=
		protocol.BuildRebootRequest()

	if err := c.Port.WriteAll(
		request,
	); err != nil {
		return fmt.Errorf(
			"enviar reboot: %w",
			err,
		)
	}

	return nil
}
