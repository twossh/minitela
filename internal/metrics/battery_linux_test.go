//go:build linux

package metrics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadBatteryFrom(t *testing.T) {
	root := t.TempDir()

	bat := filepath.Join(
		root,
		"BAT0",
	)

	if err := os.MkdirAll(
		bat,
		0o755,
	); err != nil {
		t.Fatal(err)
	}

	write := func(name, value string) {
		t.Helper()

		if err := os.WriteFile(
			filepath.Join(bat, name),
			[]byte(value+"\n"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
	}

	write("type", "Battery")
	write("capacity", "82")
	write("status", "Discharging")

	got, err := readBatteryFrom(root)
	if err != nil {
		t.Fatal(err)
	}

	if got.Capacity != 82 {
		t.Fatalf(
			"Capacity=%d esperado=82",
			got.Capacity,
		)
	}

	if got.Status != "Discharging" {
		t.Fatalf(
			"Status=%q",
			got.Status,
		)
	}
}
