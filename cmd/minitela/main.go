package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/twossh/minitela/internal/r15m"
)

const (
	appName    = "MiniTela"
	appVersion = "0.1.0-dev"
)

func main() {
	setBrightness := flag.Int(
		"set-brightness",
		-1,
		"define o brilho da MiniTela entre 0 e 100",
	)

	screenName := flag.String(
		"screen",
		"",
		"seleciona a tela: whatsapp, notes, monitor ou weather",
	)

	setPage := flag.Int(
		"set-page",
		-1,
		"define diretamente a página 1-4 (diagnóstico)",
	)

	flag.Parse()

	if *setBrightness < -1 ||
		*setBrightness > 100 {
		fmt.Fprintln(
			os.Stderr,
			"Erro: brilho deve estar entre 0 e 100.",
		)
		os.Exit(2)
	}

	if *setPage != -1 &&
		(*setPage < 1 || *setPage > 4) {
		fmt.Fprintln(
			os.Stderr,
			"Erro: página deve estar entre 1 e 4.",
		)
		os.Exit(2)
	}

	if *screenName != "" &&
		*setPage != -1 {
		fmt.Fprintln(
			os.Stderr,
			"Erro: use --screen ou --set-page, não os dois.",
		)
		os.Exit(2)
	}

	var selectedScreen r15m.Screen

	if *screenName != "" {
		var err error

		selectedScreen, err =
			r15m.ParseScreen(*screenName)

		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Erro: %v\n",
				err,
			)
			os.Exit(2)
		}
	}

	fmt.Printf(
		"%s %s\n",
		appName,
		appVersion,
	)

	fmt.Println(
		"MiniTela para Positivo R15M",
	)

	fmt.Printf(
		"Sistema: %s/%s\n",
		runtime.GOOS,
		runtime.GOARCH,
	)

	fmt.Println()
	fmt.Println("Procurando MiniTela...")

	conn, err := r15m.Connect()
	if err != nil {
		fmt.Println()
		fmt.Println("Status: não conectada")
		fmt.Printf("Erro: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println()
	fmt.Println("Positivo R15M detectado")

	fmt.Printf(
		"Dispositivo : %s\n",
		conn.Device.Path,
	)

	fmt.Printf(
		"USB         : %s:%s\n",
		conn.Device.VendorID,
		conn.Device.ProductID,
	)

	if conn.Device.Product != "" {
		fmt.Printf(
			"Produto     : %s\n",
			conn.Device.Product,
		)
	}

	if conn.Device.Serial != "" {
		fmt.Printf(
			"Serial      : %s\n",
			conn.Device.Serial,
		)
	}

	fmt.Printf(
		"Handshake   : % X\n",
		conn.HandshakeResponse,
	)

	page, err := conn.ReadRegister(
		r15m.RegisterCurrentPage,
	)

	if err != nil {
		fmt.Printf(
			"Página      : erro: %v\n",
			err,
		)
	} else {
		screen := r15m.Screen(page)

		fmt.Printf(
			"Página      : %d (%s)\n",
			page,
			screen.String(),
		)
	}

	brightness, err := conn.ReadRegister(
		r15m.RegisterBrightness,
	)

	if err != nil {
		fmt.Printf(
			"Brilho      : erro: %v\n",
			err,
		)
	} else {
		fmt.Printf(
			"Brilho      : %d%%\n",
			brightness,
		)
	}

	if *setBrightness >= 0 {
		fmt.Println()

		fmt.Printf(
			"Definindo brilho para %d%%...\n",
			*setBrightness,
		)

		actual, err :=
			conn.WriteRegisterVerified(
				r15m.RegisterBrightness,
				uint32(*setBrightness),
			)

		if err != nil {
			fmt.Printf(
				"Erro ao definir brilho: %v\n",
				err,
			)
			os.Exit(1)
		}

		fmt.Printf(
			"Brilho confirmado: %d%%\n",
			actual,
		)
	}

	if *screenName != "" {
		fmt.Println()

		fmt.Printf(
			"Selecionando tela %s...\n",
			selectedScreen.String(),
		)

		if err :=
			conn.SetScreen(
				selectedScreen,
			); err != nil {
			fmt.Printf(
				"Erro ao selecionar tela: %v\n",
				err,
			)
			os.Exit(1)
		}

		fmt.Printf(
			"Tela confirmada: %s (%d)\n",
			selectedScreen.String(),
			selectedScreen,
		)
	}

	if *setPage >= 0 {
		fmt.Println()

		fmt.Printf(
			"Definindo página %d...\n",
			*setPage,
		)

		actual, err :=
			conn.WriteRegisterVerified(
				r15m.RegisterCurrentPage,
				uint32(*setPage),
			)

		if err != nil {
			fmt.Printf(
				"Erro ao definir página: %v\n",
				err,
			)
			os.Exit(1)
		}

		fmt.Printf(
			"Página confirmada: %d (%s)\n",
			actual,
			r15m.Screen(actual).String(),
		)
	}

	fmt.Println("Status      : conectada")
}
