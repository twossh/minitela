package config

import (
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.LastScreen != "monitor" {
		t.Fatalf(
			"LastScreen=%q esperado=monitor",
			cfg.LastScreen,
		)
	}

	if !cfg.RestoreLastScreen {
		t.Fatal(
			"RestoreLastScreen deveria iniciar habilitado",
		)
	}

	if cfg.Brightness != nil {
		t.Fatal(
			"Brightness deveria iniciar sem valor",
		)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(
		t.TempDir(),
		"minitela",
		"config.json",
	)

	brightness := 47

	want := Config{
		Version:                1,
		LastScreen:             "weather",
		City:                   "Porto Alegre",
		Brightness:             &brightness,
		RestoreLastScreen:      true,
		Autostart:              true,
		StartMinimized:         true,
		MonitorIntervalSeconds: 5,
	}

	if err := SaveTo(
		path,
		want,
	); err != nil {
		t.Fatal(err)
	}

	got, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.LastScreen != want.LastScreen {
		t.Fatalf(
			"LastScreen=%q esperado=%q",
			got.LastScreen,
			want.LastScreen,
		)
	}

	if got.City != want.City {
		t.Fatalf(
			"City=%q esperado=%q",
			got.City,
			want.City,
		)
	}

	if got.Brightness == nil ||
		*got.Brightness != 47 {
		t.Fatalf(
			"Brightness inválido",
		)
	}

	if !got.Autostart {
		t.Fatal(
			"Autostart deveria estar habilitado",
		)
	}
}

func TestNormalizeInvalidValues(t *testing.T) {
	brightness := 150

	cfg := Config{
		Version:                0,
		LastScreen:             "pagina-inexistente",
		Brightness:             &brightness,
		MonitorIntervalSeconds: 0,
	}

	normalize(&cfg)

	if cfg.LastScreen != "monitor" {
		t.Fatalf(
			"LastScreen=%q",
			cfg.LastScreen,
		)
	}

	if cfg.Brightness == nil ||
		*cfg.Brightness != 100 {
		t.Fatalf(
			"Brightness não foi limitado a 100",
		)
	}

	if cfg.MonitorIntervalSeconds != 10 {
		t.Fatalf(
			"MonitorIntervalSeconds=%d",
			cfg.MonitorIntervalSeconds,
		)
	}
}
