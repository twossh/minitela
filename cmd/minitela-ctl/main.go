package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/twossh/minitela/internal/r15m"
)

const serviceName = "minitela.service"

func main() {
	on := flag.Bool(
		"on",
		false,
		"liga a MiniTela e inicia o serviço",
	)

	off := flag.Bool(
		"off",
		false,
		"desliga a MiniTela e para o serviço",
	)

	reboot := flag.Bool(
		"reboot",
		false,
		"reinicia o controlador da MiniTela",
	)

	restart := flag.Bool(
		"restart",
		false,
		"reinicia somente o serviço MiniTela",
	)

	status := flag.Bool(
		"status",
		false,
		"exibe o estado do serviço",
	)

	flag.Parse()

	count := 0

	for _, value := range []bool{
		*on,
		*off,
		*reboot,
		*restart,
		*status,
	} {
		if value {
			count++
		}
	}

	if count != 1 {
		fail(
			"use exatamente uma opção: --on, --off, --reboot, --restart ou --status",
		)
	}

	switch {
	case *on:
		powerOn()

	case *off:
		powerOff()

	case *reboot:
		rebootDevice()

	case *restart:
		if err := systemctl(
			"restart",
		); err != nil {
			fail(err.Error())
		}

		fmt.Println(
			"Serviço MiniTela reiniciado.",
		)

	case *status:
		if err := systemctl(
			"is-active",
		); err != nil {
			os.Exit(1)
		}
	}
}

func powerOn() {
	if err := systemctl(
		"start",
	); err != nil {
		fail(
			fmt.Sprintf(
				"ligar serviço: %v",
				err,
			),
		)
	}

	fmt.Println(
		"MiniTela ligada.",
	)

	fmt.Println(
		"O serviço restaurará o brilho e a última tela.",
	)
}

func powerOff() {
	// Primeiro liberamos /dev/ttyACM*.
	if err := systemctl(
		"stop",
	); err != nil {
		fail(
			fmt.Sprintf(
				"parar serviço: %v",
				err,
			),
		)
	}

	conn, err := r15m.Connect()
	if err != nil {
		fail(
			fmt.Sprintf(
				"conectar ao R15M: %v",
				err,
			),
		)
	}
	defer conn.Close()

	if err := conn.PowerOff(); err != nil {
		fail(err.Error())
	}

	fmt.Println(
		"MiniTela desligada.",
	)

	fmt.Println(
		"Brilho físico definido em 0%.",
	)

	fmt.Println(
		"Serviço MiniTela parado.",
	)
}

func rebootDevice() {
	// O serviço mantém a serial aberta, então primeiro
	// precisamos pará-lo.
	if err := systemctl(
		"stop",
	); err != nil {
		fail(
			fmt.Sprintf(
				"parar serviço: %v",
				err,
			),
		)
	}

	// Mesmo se algo falhar depois, tentamos recuperar
	// o serviço automaticamente.
	restartService := true

	defer func() {
		if restartService {
			_ = systemctl(
				"start",
			)
		}
	}()

	conn, err := r15m.Connect()
	if err != nil {
		fail(
			fmt.Sprintf(
				"conectar ao R15M: %v",
				err,
			),
		)
	}

	fmt.Println(
		"Enviando comando de reinicialização...",
	)

	err = conn.Reboot()

	_ = conn.Close()

	if err != nil {
		fail(
			fmt.Sprintf(
				"reiniciar MiniTela: %v",
				err,
			),
		)
	}

	fmt.Println(
		"Controlador reiniciado.",
	)

	// Damos tempo para o dispositivo USB desaparecer
	// e enumerar novamente.
	time.Sleep(
		3 * time.Second,
	)

	if err := systemctl(
		"start",
	); err != nil {
		fail(
			fmt.Sprintf(
				"restaurar serviço: %v",
				err,
			),
		)
	}

	restartService = false

	fmt.Println(
		"Serviço MiniTela iniciado novamente.",
	)
}

func systemctl(
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

func fail(message string) {
	fmt.Fprintln(
		os.Stderr,
		"Erro:",
		message,
	)

	os.Exit(1)
}
