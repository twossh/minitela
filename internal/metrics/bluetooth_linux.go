//go:build linux

package metrics

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type Bluetooth struct {
	Address   string
	Name      string
	Connected bool
}

// ReadBluetooth returns the first Bluetooth device currently
// connected through BlueZ.
//
// No connected device is not considered an error.
func ReadBluetooth() (*Bluetooth, error) {
	output, err := exec.Command(
		"bluetoothctl",
		"devices",
		"Connected",
	).Output()

	if err != nil {
		return nil, fmt.Errorf(
			"consultar dispositivos Bluetooth: %w",
			err,
		)
	}

	return parseBluetoothDevices(
		string(output),
	)
}

func parseBluetoothDevices(
	output string,
) (*Bluetooth, error) {
	scanner := bufio.NewScanner(
		strings.NewReader(output),
	)

	for scanner.Scan() {
		line := strings.TrimSpace(
			scanner.Text(),
		)

		if line == "" {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		if !strings.EqualFold(
			fields[0],
			"Device",
		) {
			continue
		}

		address := fields[1]

		name := ""

		if len(fields) > 2 {
			name = strings.Join(
				fields[2:],
				" ",
			)
		}

		if name == "" {
			name = address
		}

		return &Bluetooth{
			Address:   address,
			Name:      name,
			Connected: true,
		}, nil
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf(
			"ler saída do bluetoothctl: %w",
			err,
		)
	}

	return &Bluetooth{
		Connected: false,
	}, nil
}
