//go:build linux

package metrics

import "testing"

func TestParseBluetoothDevicesConnected(t *testing.T) {
	input := `
Device D2:78:C4:27:37:A8 Pebble M350s
`

	got := parseBluetoothDevices(input)

	if !got.Connected {
		t.Fatal(
			"Bluetooth deveria estar conectado",
		)
	}

	if got.Address != "D2:78:C4:27:37:A8" {
		t.Fatalf(
			"Address=%q esperado=%q",
			got.Address,
			"D2:78:C4:27:37:A8",
		)
	}

	if got.Name != "Pebble M350s" {
		t.Fatalf(
			"Name=%q esperado=%q",
			got.Name,
			"Pebble M350s",
		)
	}
}

func TestParseBluetoothDevicesDisconnected(t *testing.T) {
	got := parseBluetoothDevices("")

	if got.Connected {
		t.Fatal(
			"Bluetooth deveria estar desconectado",
		)
	}

	if got.Name != "" {
		t.Fatalf(
			"Name=%q esperado vazio",
			got.Name,
		)
	}

	if got.Address != "" {
		t.Fatalf(
			"Address=%q esperado vazio",
			got.Address,
		)
	}
}

func TestParseBluetoothDevicesMultiple(t *testing.T) {
	input := `
Device D2:78:C4:27:37:A8 Pebble M350s
Device AA:BB:CC:DD:EE:FF Outro Dispositivo
`

	got := parseBluetoothDevices(input)

	if !got.Connected {
		t.Fatal(
			"Bluetooth deveria estar conectado",
		)
	}

	if got.Name != "Pebble M350s" {
		t.Fatalf(
			"primeiro dispositivo=%q esperado=%q",
			got.Name,
			"Pebble M350s",
		)
	}
}

func TestParseBluetoothDeviceWithoutName(t *testing.T) {
	input := `
Device AA:BB:CC:DD:EE:FF
`

	got := parseBluetoothDevices(input)

	if !got.Connected {
		t.Fatal(
			"Bluetooth deveria estar conectado",
		)
	}

	if got.Name != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf(
			"Name=%q esperado endereço",
			got.Name,
		)
	}
}
