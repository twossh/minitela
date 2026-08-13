//go:build linux

package device

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsR15MUSB(t *testing.T) {
	tests := []struct {
		vendor  string
		product string
		name    string
		want    bool
	}{
		{"0324", "0324", "", true},
		{"0324", "0324", "sxw-admin_CDC_DEMO", true},
		{"ffff", "ffff", "sxw-admin_CDC_DEMO", true},
		{"1234", "5678", "Other Device", false},
	}

	for _, tt := range tests {
		got := isR15MUSB(tt.vendor, tt.product, tt.name)

		if got != tt.want {
			t.Errorf(
				"isR15MUSB(%q,%q,%q) = %v, want %v",
				tt.vendor,
				tt.product,
				tt.name,
				got,
				tt.want,
			)
		}
	}
}

func TestDetectR15MFromFakeSysfs(t *testing.T) {
	root := t.TempDir()

	classTTY := filepath.Join(root, "sys", "class", "tty")
	devRoot := filepath.Join(root, "dev")

	if err := os.MkdirAll(classTTY, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(devRoot, 0o755); err != nil {
		t.Fatal(err)
	}

	usbDevice := filepath.Join(
		root,
		"sys",
		"devices",
		"pci0000:00",
		"usb1",
		"1-1",
	)

	ttyDir := filepath.Join(
		usbDevice,
		"1-1:1.0",
		"tty",
		"ttyACM0",
	)

	if err := os.MkdirAll(ttyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	write := func(name, value string) {
		t.Helper()

		if err := os.WriteFile(
			filepath.Join(usbDevice, name),
			[]byte(value+"\n"),
			0o644,
		); err != nil {
			t.Fatal(err)
		}
	}

	write("idVendor", "0324")
	write("idProduct", "0324")
	write("product", "sxw-admin_CDC_DEMO")
	write("serial", "CherryUSB_test")

	link := filepath.Join(classTTY, "ttyACM0")

	if err := os.Symlink(ttyDir, link); err != nil {
		t.Fatal(err)
	}

	got, err := detectR15M(classTTY, devRoot)
	if err != nil {
		t.Fatal(err)
	}

	if got.TTY != "ttyACM0" {
		t.Fatalf("TTY = %q, want ttyACM0", got.TTY)
	}

	if got.VendorID != "0324" {
		t.Fatalf("VendorID = %q, want 0324", got.VendorID)
	}

	if got.ProductID != "0324" {
		t.Fatalf("ProductID = %q, want 0324", got.ProductID)
	}

	if got.Product != "sxw-admin_CDC_DEMO" {
		t.Fatalf(
			"Product = %q, want sxw-admin_CDC_DEMO",
			got.Product,
		)
	}

	if got.Serial != "CherryUSB_test" {
		t.Fatalf(
			"Serial = %q, want CherryUSB_test",
			got.Serial,
		)
	}
}
