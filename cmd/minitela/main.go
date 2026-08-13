package main

import (
	"fmt"
	"runtime"
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
	fmt.Println("Inicialização concluída.")
}
