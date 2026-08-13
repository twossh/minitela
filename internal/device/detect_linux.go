//go:build linux

package device

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	r15mVendorID  = "0324"
	r15mProductID = "0324"
)

// Device describes a MiniTela-compatible serial device detected by Linux.
type Device struct {
	Path      string
	TTY       string
	VendorID  string
	ProductID string
	Product   string
	Serial    string
	SysPath   string
}

// DetectR15M searches Linux sysfs for the USB CDC serial device used by
// the Positivo Vision R15M auxiliary display.
func DetectR15M() (*Device, error) {
	return detectR15M("/sys/class/tty", "/dev")
}

func detectR15M(sysClassTTY, devRoot string) (*Device, error) {
	entries, err := os.ReadDir(sysClassTTY)
	if err != nil {
		return nil, fmt.Errorf("read sysfs tty directory: %w", err)
	}

	var candidates []string

	for _, entry := range entries {
		name := entry.Name()

		if !strings.HasPrefix(name, "ttyACM") {
			continue
		}

		candidates = append(candidates, name)

		classPath := filepath.Join(sysClassTTY, name)

		realPath, err := filepath.EvalSymlinks(classPath)
		if err != nil {
			continue
		}

		attrs, usbPath, err := findUSBAttributes(realPath)
		if err != nil {
			continue
		}

		if !isR15MUSB(attrs.vendorID, attrs.productID, attrs.product) {
			continue
		}

		return &Device{
			Path:      filepath.Join(devRoot, name),
			TTY:       name,
			VendorID:  attrs.vendorID,
			ProductID: attrs.productID,
			Product:   attrs.product,
			Serial:    attrs.serial,
			SysPath:   usbPath,
		}, nil
	}

	if len(candidates) == 0 {
		return nil, errors.New("nenhum dispositivo ttyACM encontrado")
	}

	return nil, fmt.Errorf(
		"nenhum Positivo R15M encontrado entre: %s",
		strings.Join(candidates, ", "),
	)
}

type usbAttributes struct {
	vendorID  string
	productID string
	product   string
	serial    string
}

func findUSBAttributes(start string) (usbAttributes, string, error) {
	current := filepath.Clean(start)

	for {
		vendor, vendorErr := readAttribute(
			filepath.Join(current, "idVendor"),
		)
		productID, productErr := readAttribute(
			filepath.Join(current, "idProduct"),
		)

		if vendorErr == nil && productErr == nil {
			product, _ := readAttribute(
				filepath.Join(current, "product"),
			)
			serial, _ := readAttribute(
				filepath.Join(current, "serial"),
			)

			return usbAttributes{
				vendorID:  strings.ToLower(vendor),
				productID: strings.ToLower(productID),
				product:   product,
				serial:    serial,
			}, current, nil
		}

		parent := filepath.Dir(current)

		if parent == current {
			break
		}

		current = parent
	}

	return usbAttributes{}, "", errors.New(
		"USB attributes not found in sysfs hierarchy",
	)
}

func readAttribute(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func isR15MUSB(vendorID, productID, product string) bool {
	if strings.EqualFold(vendorID, r15mVendorID) &&
		strings.EqualFold(productID, r15mProductID) {
		return true
	}

	// Secondary identification for known CherryUSB firmware.
	return strings.Contains(
		strings.ToLower(product),
		"sxw-admin_cdc_demo",
	)
}
