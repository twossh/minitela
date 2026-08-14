package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/weather"
)

func main() {
	cfg, err :=
		config.Load()

	if err != nil {
		fail(err)
	}

	if cfg.City == "" {
		fail(fmt.Errorf(
			"cidade não configurada",
		))
	}

	if cfg.WeatherAPIKey == "" {
		fail(fmt.Errorf(
			"WeatherAPI não configurada",
		))
	}

	fmt.Println(
		"=== MiniTela WeatherAPI Test ===",
	)

	fmt.Printf(
		"Cidade configurada: %s\n",
		cfg.City,
	)

	fmt.Println(
		"WeatherAPI: configurada",
	)

	fmt.Println()
	fmt.Println(
		"Consultando...",
	)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			15*time.Second,
		)

	defer cancel()

	client :=
		weather.NewClient(
			cfg.WeatherAPIKey,
		)

	forecast, err :=
		client.GetForecast(
			ctx,
			cfg.City,
		)

	if err != nil {
		fail(err)
	}

	fmt.Println()
	fmt.Printf(
		"Local: %s - %s - %s\n",
		forecast.Location.Name,
		forecast.Location.Region,
		forecast.Location.Country,
	)

	for i, day := range forecast.Days {
		fmt.Printf(
			"Dia %d: %s | %s | %d°/%d° | icon=%d\n",
			i+1,
			day.Date.Format(
				"02/01/2006",
			),
			day.Condition,
			day.MinTemp,
			day.MaxTemp,
			weather.R15MIcon(
				day.Condition,
			),
		)
	}

	fmt.Println()
	fmt.Println(
		"WeatherAPI: OK",
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
