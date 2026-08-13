package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/twossh/minitela/internal/device"
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

	dev, err := device.DetectR15M()
	if err != nil {
		fmt.Println()
		fmt.Println("Status: não conectada")
		fmt.Printf("Detalhes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Positivo R15M detectado")
	fmt.Printf("Dispositivo : %s\n", dev.Path)
	fmt.Printf("USB         : %s:%s\n", dev.VendorID, dev.ProductID)

	if dev.Product != "" {
		fmt.Printf("Produto     : %s\n", dev.Product)
	}

	if dev.Serial != "" {
		fmt.Printf("Serial      : %s\n", dev.Serial)
	}

	fmt.Println("Status      : detectada")
}
