package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const currentVersion = 1

type Config struct {
	Version int `json:"version"`

	LastScreen string `json:"last_screen"`
	City       string `json:"city"`

	WeatherAPIKey string `json:"weather_api_key,omitempty"`

	Brightness *int `json:"brightness,omitempty"`

	RestoreLastScreen bool `json:"restore_last_screen"`
	Autostart         bool `json:"autostart"`
	StartMinimized    bool `json:"start_minimized"`

	MonitorIntervalSeconds int `json:"monitor_interval_seconds"`
}

func Default() Config {
	return Config{
		Version:                currentVersion,
		LastScreen:             "monitor",
		City:                   "",
		WeatherAPIKey:          "",
		Brightness:             nil,
		RestoreLastScreen:      true,
		Autostart:              false,
		StartMinimized:         true,
		MonitorIntervalSeconds: 10,
	}
}

func Path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf(
			"localizar diretório de configuração: %w",
			err,
		)
	}

	return filepath.Join(
		base,
		"minitela",
		"config.json",
	), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}

	return LoadFrom(path)
}

func LoadFrom(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)

	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}

	if err != nil {
		return Config{}, fmt.Errorf(
			"ler configuração: %w",
			err,
		)
	}

	if err := json.Unmarshal(
		data,
		&cfg,
	); err != nil {
		return Config{}, fmt.Errorf(
			"configuração inválida: %w",
			err,
		)
	}

	normalize(&cfg)

	return cfg, nil
}

func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}

	return SaveTo(path, cfg)
}

func SaveTo(
	path string,
	cfg Config,
) error {
	normalize(&cfg)

	dir := filepath.Dir(path)

	if err := os.MkdirAll(
		dir,
		0o700,
	); err != nil {
		return fmt.Errorf(
			"criar diretório de configuração: %w",
			err,
		)
	}

	data, err := json.MarshalIndent(
		cfg,
		"",
		"  ",
	)
	if err != nil {
		return fmt.Errorf(
			"codificar configuração: %w",
			err,
		)
	}

	data = append(data, '\n')

	tmp := path + ".tmp"

	if err := os.WriteFile(
		tmp,
		data,
		0o600,
	); err != nil {
		return fmt.Errorf(
			"gravar configuração temporária: %w",
			err,
		)
	}

	if err := os.Rename(
		tmp,
		path,
	); err != nil {
		_ = os.Remove(tmp)

		return fmt.Errorf(
			"salvar configuração: %w",
			err,
		)
	}

	return nil
}

func normalize(cfg *Config) {
	if cfg.Version <= 0 {
		cfg.Version = currentVersion
	}

	switch cfg.LastScreen {
	case "whatsapp",
		"notes",
		"monitor",
		"weather":
	default:
		cfg.LastScreen = "monitor"
	}

	cfg.City =
		strings.TrimSpace(cfg.City)

	cfg.WeatherAPIKey =
		strings.TrimSpace(cfg.WeatherAPIKey)

	if cfg.MonitorIntervalSeconds < 1 {
		cfg.MonitorIntervalSeconds = 10
	}

	if cfg.Brightness != nil {
		value := *cfg.Brightness

		if value < 0 {
			value = 0
		}

		if value > 100 {
			value = 100
		}

		cfg.Brightness = &value
	}
}
