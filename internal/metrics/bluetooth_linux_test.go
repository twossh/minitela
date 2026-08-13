//go:build linux

package metrics

import "testing"

func TestParseBluetoothDevice(t *testing.T) {
	input :=
		"Device AA:BB:CC:DD:EE:FF Pebble M350s\n"

	got, err := parseBluetoothDevices(input)
	if err != nil {
		t.Fatal(err)
	}

	if !got.Connected {
		t.Fatal("dispositivo deveria estar conectado")
	}

	if got.Address != "AA:BB:CC:DD:EE:FF" {
		t.Fatalf(
			"Address=%q",
			got.Address,
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

func TestParseBluetoothNoDevice(t *testing.T) {
	got, err := parseBluetoothDevices("")
	if err != nil {
		t.Fatal(err)
	}

	if got.Connected {
		t.Fatal(
			"não deveria existir dispositivo conectado",
		)
	}
}
