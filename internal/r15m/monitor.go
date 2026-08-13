package r15m

import (
	"fmt"

	"github.com/twossh/minitela/internal/metrics"
)

type BatterySyncResult struct {
	Capacity int
	Level    uint32
	Status   string
}

type WiFiSyncResult struct {
	Interface string
	SSID      string
	Display   string
	SignalDBM int
	Quality   int
}

func (c *Connection) SyncBattery() (
	*BatterySyncResult,
	error,
) {
	battery, err := metrics.ReadBattery()
	if err != nil {
		return nil, err
	}

	text := BatteryText(
		battery.Capacity,
	)

	level := BatteryLevel(
		battery.Capacity,
	)

	if err := c.WriteStringRegister(
		RegisterBatteryText,
		text,
	); err != nil {
		return nil, fmt.Errorf(
			"atualizar percentual da bateria: %w",
			err,
		)
	}

	// Register 1150 is a display-state register.
	// The R15M ACK is sufficient; reading it back does not
	// reliably return the value just written.
	if err := c.WriteRegister(
		RegisterBatteryLevel,
		level,
	); err != nil {
		return nil, fmt.Errorf(
			"atualizar nível gráfico da bateria: %w",
			err,
		)
	}

	return &BatterySyncResult{
		Capacity: battery.Capacity,
		Level:    level,
		Status:   battery.Status,
	}, nil
}

func (c *Connection) SyncWiFi() (
	*WiFiSyncResult,
	error,
) {
	wifi, err := metrics.ReadWiFi()
	if err != nil {
		return nil, err
	}

	displaySSID := DisplayText(
		wifi.SSID,
	)

	if err := c.WriteStringRegister(
		RegisterWiFiSSID,
		[]byte(displaySSID),
	); err != nil {
		return nil, fmt.Errorf(
			"atualizar SSID: %w",
			err,
		)
	}

	if err := c.WriteRegister(
		RegisterWiFiQuality,
		uint32(wifi.Quality),
	); err != nil {
		return nil, fmt.Errorf(
			"atualizar qualidade Wi-Fi: %w",
			err,
		)
	}

	return &WiFiSyncResult{
		Interface: wifi.Interface,
		SSID:      wifi.SSID,
		Display:   displaySSID,
		SignalDBM: wifi.SignalDBM,
		Quality:   wifi.Quality,
	}, nil
}
