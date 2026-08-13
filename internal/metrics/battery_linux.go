//go:build linux

package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Battery struct {
	Name     string
	Capacity int
	Status   string
	Path     string
}

func ReadBattery() (*Battery, error) {
	return readBatteryFrom("/sys/class/power_supply")
}

func readBatteryFrom(base string) (*Battery, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf(
			"ler power_supply: %w",
			err,
		)
	}

	for _, entry := range entries {
		path := filepath.Join(
			base,
			entry.Name(),
		)

		typeData, err := os.ReadFile(
			filepath.Join(path, "type"),
		)
		if err != nil {
			continue
		}

		if strings.TrimSpace(
			string(typeData),
		) != "Battery" {
			continue
		}

		capacityData, err := os.ReadFile(
			filepath.Join(path, "capacity"),
		)
		if err != nil {
			continue
		}

		capacity, err := strconv.Atoi(
			strings.TrimSpace(
				string(capacityData),
			),
		)
		if err != nil {
			continue
		}

		if capacity < 0 {
			capacity = 0
		}

		if capacity > 100 {
			capacity = 100
		}

		statusData, _ := os.ReadFile(
			filepath.Join(path, "status"),
		)

		return &Battery{
			Name:     entry.Name(),
			Capacity: capacity,
			Status: strings.TrimSpace(
				string(statusData),
			),
			Path: path,
		}, nil
	}

	return nil, fmt.Errorf(
		"nenhuma bateria encontrada",
	)
}
