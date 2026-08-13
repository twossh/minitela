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

	flag.Parse()

	if *setBrightness < -1 || *setBrightness > 100 {
		fmt.Fprintln(
			os.Stderr,
			"Erro: brilho deve estar entre 0 e 100.",
		)
		os.Exit(2)
	}

	fmt.Printf("%s %s\n", appName, appVersion)
	fmt.Println("MiniTela para Positivo R15M")
	fmt.Printf("Sistema: %s/%s\n", runtime.GOOS, runtime.GOARCH)
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
	fmt.Printf("Dispositivo : %s\n", conn.Device.Path)
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
		fmt.Printf(
			"Página      : %d\n",
			page,
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

		actual, err := conn.WriteRegisterVerified(
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

	fmt.Println("Status      : conectada")
}
