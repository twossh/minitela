//go:build linux

package metrics

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Bluetooth struct {
	Connected bool
	Address   string
	Name      string
}

func ReadBluetooth() (*Bluetooth, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"bluetoothctl",
		"devices",
		"Connected",
	)

	output, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf(
			"timeout consultando Bluetooth",
		)
	}

	if err != nil {
		// Para o Monitor, Bluetooth desligado ou sem
		// controlador não deve derrubar toda a aplicação.
		return &Bluetooth{
			Connected: false,
		}, nil
	}

	return parseBluetoothDevices(
		string(output),
	), nil
}

// parseBluetoothDevices interprets the output of:
//
//	bluetoothctl devices Connected
//
// Expected format:
//
//	Device D2:78:C4:27:37:A8 Pebble M350s
//
// If multiple devices are connected, the first one is used by
// the current R15M Monitor layout because it has only one field
// available for a Bluetooth device name.
func parseBluetoothDevices(
	output string,
) *Bluetooth {
	scanner := bufio.NewScanner(
		strings.NewReader(output),
	)

	for scanner.Scan() {
		line := strings.TrimSpace(
			scanner.Text(),
		)

		if !strings.HasPrefix(
			line,
			"Device ",
		) {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		address := fields[1]

		name := ""

		if len(fields) >= 3 {
			name = strings.TrimSpace(
				strings.Join(
					fields[2:],
					" ",
				),
			)
		}

		if name == "" {
			name = address
		}

		return &Bluetooth{
			Connected: true,
			Address:   address,
			Name:      name,
		}
	}

	return &Bluetooth{
		Connected: false,
	}
}
