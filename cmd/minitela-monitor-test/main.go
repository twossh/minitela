package main

import (
	"fmt"
	"os"

	"github.com/twossh/minitela/internal/metrics"
	"github.com/twossh/minitela/internal/r15m"
)

func main() {
	battery, err := metrics.ReadBattery()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Bateria Linux: %v\n",
			err,
		)
		os.Exit(1)
	}

	fmt.Println("Bateria detectada:")
	fmt.Printf(
		"  Dispositivo : %s\n",
		battery.Name,
	)
	fmt.Printf(
		"  Capacidade  : %d%%\n",
		battery.Capacity,
	)
	fmt.Printf(
		"  Status      : %s\n",
		battery.Status,
	)
	fmt.Printf(
		"  Nível R15M  : %d\n",
		r15m.BatteryLevel(
			battery.Capacity,
		),
	)

	fmt.Println()
	fmt.Println("Conectando à MiniTela...")

	conn, err := r15m.Connect()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Conexão: %v\n",
			err,
		)
		os.Exit(1)
	}
	defer conn.Close()

	if err := conn.SetScreen(
		r15m.ScreenMonitor,
	); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Selecionar Monitor: %v\n",
			err,
		)
		os.Exit(1)
	}

	result, err := conn.SyncBattery()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Sincronizar bateria: %v\n",
			err,
		)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("MiniTela atualizada:")
	fmt.Printf(
		"  Bateria : %d%%\n",
		result.Capacity,
	)
	fmt.Printf(
		"  Nível   : %d/3\n",
		result.Level,
	)
	fmt.Printf(
		"  Status  : %s\n",
		result.Status,
	)

	fmt.Println()
	fmt.Println("OK")
}
