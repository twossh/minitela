package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/twossh/minitela/internal/protocol"
	"github.com/twossh/minitela/internal/r15m"
)

const serviceName = "minitela.service"

func main() {
	fmt.Println(
		"=== MiniTela - Upload Protocol Probe ===",
	)

	fmt.Println()

	fmt.Println(
		"Modo somente leitura.",
	)

	fmt.Println(
		"Nenhum arquivo será enviado.",
	)

	fmt.Println()

	wasActive :=
		serviceIsActive()

	if wasActive {
		fmt.Println(
			"Parando temporariamente minitela.service...",
		)

		if err :=
			serviceAction(
				"stop",
			); err != nil {
			fail(
				fmt.Errorf(
					"parar serviço: %w",
					err,
				),
			)
		}

		defer func() {
			fmt.Println()
			fmt.Println(
				"Restaurando minitela.service...",
			)

			if err :=
				serviceAction(
					"start",
				); err != nil {
				fmt.Fprintf(
					os.Stderr,
					"AVISO: não foi possível restaurar o serviço: %v\n",
					err,
				)
				return
			}

			fmt.Println(
				"Serviço restaurado.",
			)
		}()
	}

	fmt.Println(
		"Conectando ao R15M...",
	)

	conn, err :=
		r15m.Connect()

	if err != nil {
		fail(
			fmt.Errorf(
				"conectar: %w",
				err,
			),
		)
	}

	fmt.Printf(
		"Dispositivo : %s\n",
		conn.Device.Path,
	)

	fmt.Println(
		"Handshake   : OK",
	)

	fmt.Println()
	fmt.Println(
		"Consultando GetDownloadStatus...",
	)

	status, err :=
		conn.ProbeDownloadStatus()

	_ = conn.Close()

	if err != nil {
		fail(err)
	}

	fmt.Println()
	fmt.Println(
		"=== RESPOSTA DO R15M ===",
	)

	fmt.Printf(
		"Status      : 0x%02X (%s)\n",
		status.Status,
		statusName(
			status.Status,
		),
	)

	fmt.Printf(
		"File ID/MD5 : %x\n",
		status.FileID,
	)

	fmt.Printf(
		"Offset      : %d bytes (0x%08X)\n",
		status.Offset,
		status.Offset,
	)

	fmt.Println()
	fmt.Println(
		"Probe       : OK",
	)
}

func statusName(
	status uint8,
) string {
	switch status {
	case protocol.DownloadStatePreparing:
		return "preparando download"

	case protocol.DownloadStateActive:
		return "download ativo"

	case protocol.DownloadStateAHMI:
		return "AHMI / operação normal"

	default:
		return "estado desconhecido"
	}
}

func serviceIsActive() bool {
	cmd := exec.Command(
		"systemctl",
		"--user",
		"is-active",
		"--quiet",
		serviceName,
	)

	return cmd.Run() == nil
}

func serviceAction(
	action string,
) error {
	cmd := exec.Command(
		"systemctl",
		"--user",
		action,
		serviceName,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func fail(err error) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %v\n",
		err,
	)

	os.Exit(1)
}
