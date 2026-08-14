package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/r15m"
)

func main() {
	city := flag.String(
		"city",
		"",
		"cidade utilizada pelo clima",
	)

	screen := flag.String(
		"screen",
		"",
		"última tela: whatsapp, notes, monitor ou weather",
	)

	brightness := flag.Int(
		"brightness",
		-1,
		"brilho entre 0 e 100",
	)

	weatherKey := flag.String(
		"weather-key",
		"",
		"chave WeatherAPI",
	)

	weatherKeyStdin := flag.Bool(
		"weather-key-stdin",
		false,
		"lê a chave WeatherAPI pela entrada padrão",
	)

	clearWeatherKey := flag.Bool(
		"clear-weather-key",
		false,
		"remove a chave WeatherAPI salva",
	)

	autostart := flag.String(
		"autostart",
		"",
		"on ou off",
	)

	minimized := flag.String(
		"minimized",
		"",
		"on ou off",
	)

	restore := flag.String(
		"restore",
		"",
		"on ou off",
	)

	show := flag.Bool(
		"show",
		false,
		"exibe a configuração",
	)

	flag.Parse()

	if *weatherKey != "" &&
		*weatherKeyStdin {
		fail(fmt.Errorf(
			"use --weather-key ou --weather-key-stdin, não os dois",
		))
	}

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}

	changed := false

	if *city != "" {
		cfg.City =
			strings.TrimSpace(*city)

		changed = true
	}

	if *screen != "" {
		s, err :=
			r15m.ParseScreen(*screen)

		if err != nil {
			fail(err)
		}

		cfg.LastScreen =
			screenConfigName(s)

		changed = true
	}

	if *brightness >= 0 {
		if *brightness > 100 {
			fail(fmt.Errorf(
				"brilho deve estar entre 0 e 100",
			))
		}

		value := *brightness

		cfg.Brightness = &value

		changed = true
	}

	if *weatherKey != "" {
		cfg.WeatherAPIKey =
			strings.TrimSpace(*weatherKey)

		changed = true
	}

	if *weatherKeyStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fail(fmt.Errorf(
				"ler chave WeatherAPI: %w",
				err,
			))
		}

		value :=
			strings.TrimSpace(
				string(data),
			)

		if value == "" {
			fail(fmt.Errorf(
				"chave WeatherAPI vazia",
			))
		}

		cfg.WeatherAPIKey = value

		changed = true
	}

	if *clearWeatherKey {
		cfg.WeatherAPIKey = ""
		changed = true
	}

	if *autostart != "" {
		value, err :=
			parseBool(*autostart)

		if err != nil {
			fail(err)
		}

		cfg.Autostart = value

		changed = true
	}

	if *minimized != "" {
		value, err :=
			parseBool(*minimized)

		if err != nil {
			fail(err)
		}

		cfg.StartMinimized = value

		changed = true
	}

	if *restore != "" {
		value, err :=
			parseBool(*restore)

		if err != nil {
			fail(err)
		}

		cfg.RestoreLastScreen = value

		changed = true
	}

	if changed {
		if err := config.Save(cfg); err != nil {
			fail(err)
		}
	}

	if changed || *show {
		showConfig(cfg)
	}
}

func showConfig(cfg config.Config) {
	path, _ := config.Path()

	fmt.Printf(
		"Arquivo      : %s\n",
		path,
	)

	fmt.Printf(
		"Última tela  : %s\n",
		cfg.LastScreen,
	)

	fmt.Printf(
		"Cidade       : %s\n",
		cfg.City,
	)

	if cfg.Brightness == nil {
		fmt.Println(
			"Brilho       : automático/não definido",
		)
	} else {
		fmt.Printf(
			"Brilho       : %d%%\n",
			*cfg.Brightness,
		)
	}

	if cfg.WeatherAPIKey == "" {
		fmt.Println(
			"WeatherAPI   : não configurada",
		)
	} else {
		fmt.Println(
			"WeatherAPI   : configurada",
		)
	}

	fmt.Printf(
		"Restaurar    : %t\n",
		cfg.RestoreLastScreen,
	)

	fmt.Printf(
		"Autostart    : %t\n",
		cfg.Autostart,
	)

	fmt.Printf(
		"Minimizado   : %t\n",
		cfg.StartMinimized,
	)

	fmt.Printf(
		"Intervalo    : %ds\n",
		cfg.MonitorIntervalSeconds,
	)
}

func screenConfigName(
	screen r15m.Screen,
) string {
	switch screen {
	case r15m.ScreenWhatsApp:
		return "whatsapp"

	case r15m.ScreenNotes:
		return "notes"

	case r15m.ScreenMonitor:
		return "monitor"

	case r15m.ScreenWeather:
		return "weather"

	default:
		return "monitor"
	}
}

func parseBool(
	value string,
) (bool, error) {
	switch strings.ToLower(
		strings.TrimSpace(value),
	) {
	case "on",
		"true",
		"1",
		"yes",
		"sim":
		return true, nil

	case "off",
		"false",
		"0",
		"no",
		"nao",
		"não":
		return false, nil
	}

	return false, fmt.Errorf(
		"valor inválido %q; use on ou off",
		value,
	)
}

func fail(err error) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %v\n",
		err,
	)

	os.Exit(1)
}
