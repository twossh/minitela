package main

import (
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

	fmt.Println("Status      : conectada")
}
