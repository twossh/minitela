package main

import (
	"fmt"
	"os"

	"github.com/twossh/minitela/internal/metrics"
	"github.com/twossh/minitela/internal/r15m"
)

func main() {
	fmt.Println("=== MiniTela Monitor Test ===")
	fmt.Println()

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

	wifi, err := metrics.ReadWiFi()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Wi-Fi Linux: %v\n",
			err,
		)
		os.Exit(1)
	}

	fmt.Println("Wi-Fi detectado:")
	fmt.Printf(
		"  Interface   : %s\n",
		wifi.Interface,
	)
	fmt.Printf(
		"  SSID        : %s\n",
		wifi.SSID,
	)
	fmt.Printf(
		"  Sinal       : %d dBm\n",
		wifi.SignalDBM,
	)
	fmt.Printf(
		"  Qualidade   : %d%%\n",
		wifi.Quality,
	)
	fmt.Printf(
		"  Texto R15M  : %s\n",
		r15m.DisplayText(
			wifi.SSID,
		),
	)

	fmt.Println()

	bt, err := metrics.ReadBluetooth()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Bluetooth Linux: %v\n",
			err,
		)
		os.Exit(1)
	}

	fmt.Println("Bluetooth detectado:")

	if bt.Connected {
		fmt.Println(
			"  Conectado   : sim",
		)
		fmt.Printf(
			"  Endereço    : %s\n",
			bt.Address,
		)
		fmt.Printf(
			"  Nome        : %s\n",
			bt.Name,
		)
		fmt.Printf(
			"  Texto R15M  : %s\n",
			r15m.DisplayText(
				bt.Name,
			),
		)
	} else {
		fmt.Println(
			"  Conectado   : não",
		)
	}

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

	batteryResult, err :=
		conn.SyncBattery()

	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Sincronizar bateria: %v\n",
			err,
		)
		os.Exit(1)
	}

	wifiResult, err :=
		conn.SyncWiFi()

	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Sincronizar Wi-Fi: %v\n",
			err,
		)
		os.Exit(1)
	}

	btResult, err :=
		conn.SyncBluetooth()

	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Sincronizar Bluetooth: %v\n",
			err,
		)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("MiniTela atualizada:")

	fmt.Printf(
		"  Bateria     : %d%%\n",
		batteryResult.Capacity,
	)

	fmt.Printf(
		"  Nível       : %d/3\n",
		batteryResult.Level,
	)

	fmt.Printf(
		"  Wi-Fi       : %s\n",
		wifiResult.Display,
	)

	fmt.Printf(
		"  Qualidade   : %d%%\n",
		wifiResult.Quality,
	)

	if btResult.Connected {
		fmt.Printf(
			"  Bluetooth   : %s\n",
			btResult.Display,
		)
	} else {
		fmt.Println(
			"  Bluetooth   : desconectado",
		)
	}

	fmt.Println()
	fmt.Println("OK")
}
