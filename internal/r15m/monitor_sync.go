package r15m

import (
	"fmt"

	"github.com/twossh/minitela/internal/metrics"
)

type MonitorSyncer struct {
	conn  *Connection
	cache *RegisterCache
}

type MonitorSnapshot struct {
	BatteryPercent int
	BatteryLevel   uint32
	BatteryStatus  string

	WiFiConnected bool
	WiFiSSID      string
	WiFiDisplay   string
	WiFiQuality   int
	WiFiSignalDBM int

	BluetoothConnected bool
	BluetoothName      string
	BluetoothDisplay   string

	Writes int
}

func NewMonitorSyncer(
	conn *Connection,
) *MonitorSyncer {
	return &MonitorSyncer{
		conn:  conn,
		cache: NewRegisterCache(),
	}
}

func (s *MonitorSyncer) ResetCache() {
	if s == nil || s.cache == nil {
		return
	}

	s.cache.Reset()
}

func (s *MonitorSyncer) Sync() (
	*MonitorSnapshot,
	error,
) {
	if s == nil ||
		s.conn == nil ||
		s.cache == nil {
		return nil, fmt.Errorf(
			"sincronizador Monitor inválido",
		)
	}

	result := &MonitorSnapshot{}

	if err := s.syncBattery(result); err != nil {
		return result, err
	}

	if err := s.syncWiFi(result); err != nil {
		return result, err
	}

	if err := s.syncBluetooth(result); err != nil {
		return result, err
	}

	return result, nil
}

func (s *MonitorSyncer) syncBattery(
	result *MonitorSnapshot,
) error {
	battery, err := metrics.ReadBattery()
	if err != nil {
		return fmt.Errorf(
			"bateria: %w",
			err,
		)
	}

	result.BatteryPercent = battery.Capacity
	result.BatteryStatus = battery.Status
	result.BatteryLevel =
		BatteryLevel(battery.Capacity)

	changed, err :=
		s.cache.WriteStringIfChanged(
			s.conn,
			RegisterBatteryText,
			string(
				BatteryText(
					battery.Capacity,
				),
			),
		)

	if err != nil {
		return fmt.Errorf(
			"percentual da bateria: %w",
			err,
		)
	}

	if changed {
		result.Writes++
	}

	changed, err =
		s.cache.WriteNumIfChanged(
			s.conn,
			RegisterBatteryLevel,
			result.BatteryLevel,
		)

	if err != nil {
		return fmt.Errorf(
			"nível gráfico da bateria: %w",
			err,
		)
	}

	if changed {
		result.Writes++
	}

	return nil
}

func (s *MonitorSyncer) syncWiFi(
	result *MonitorSnapshot,
) error {
	wifi, err := metrics.ReadWiFi()

	if err != nil {
		// Wi-Fi desconectado não encerra o Monitor.
		result.WiFiConnected = false

		changed, writeErr :=
			s.cache.WriteStringIfChanged(
				s.conn,
				RegisterWiFiSSID,
				" ",
			)

		if writeErr != nil {
			return fmt.Errorf(
				"limpar SSID: %w",
				writeErr,
			)
		}

		if changed {
			result.Writes++
		}

		changed, writeErr =
			s.cache.WriteNumIfChanged(
				s.conn,
				RegisterWiFiQuality,
				0,
			)

		if writeErr != nil {
			return fmt.Errorf(
				"limpar qualidade Wi-Fi: %w",
				writeErr,
			)
		}

		if changed {
			result.Writes++
		}

		return nil
	}

	result.WiFiConnected = true
	result.WiFiSSID = wifi.SSID
	result.WiFiDisplay =
		DisplayText(wifi.SSID)
	result.WiFiQuality = wifi.Quality
	result.WiFiSignalDBM = wifi.SignalDBM

	changed, err :=
		s.cache.WriteStringIfChanged(
			s.conn,
			RegisterWiFiSSID,
			result.WiFiDisplay,
		)

	if err != nil {
		return fmt.Errorf(
			"SSID: %w",
			err,
		)
	}

	if changed {
		result.Writes++
	}

	changed, err =
		s.cache.WriteNumIfChanged(
			s.conn,
			RegisterWiFiQuality,
			uint32(wifi.Quality),
		)

	if err != nil {
		return fmt.Errorf(
			"qualidade Wi-Fi: %w",
			err,
		)
	}

	if changed {
		result.Writes++
	}

	return nil
}

func (s *MonitorSyncer) syncBluetooth(
	result *MonitorSnapshot,
) error {
	bt, err := metrics.ReadBluetooth()
	if err != nil {
		return fmt.Errorf(
			"Bluetooth: %w",
			err,
		)
	}

	result.BluetoothConnected =
		bt.Connected

	display := " "

	if bt.Connected {
		result.BluetoothName = bt.Name

		display = DisplayText(
			bt.Name,
		)
	}

	result.BluetoothDisplay = display

	changed, err :=
		s.cache.WriteStringIfChanged(
			s.conn,
			RegisterBluetoothName,
			display,
		)

	if err != nil {
		return fmt.Errorf(
			"nome Bluetooth: %w",
			err,
		)
	}

	if changed {
		result.Writes++
	}

	return nil
}
