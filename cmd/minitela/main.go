package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/twossh/minitela/internal/app"
	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/r15m"
)

const (
	appName = "MiniTela"

	appVersion = "0.1.0-dev"
)

func main() {
	screenFlag :=
		flag.String(
			"screen",
			"",
			"seleciona e salva: whatsapp, notes, monitor ou weather",
		)

	brightnessFlag :=
		flag.Int(
			"set-brightness",
			-1,
			"define e salva brilho entre 0 e 100",
		)

	noRestore :=
		flag.Bool(
			"no-restore",
			false,
			"não restaura a última tela",
		)

	once :=
		flag.Bool(
			"once",
			false,
			"sincroniza uma vez e encerra",
		)

	flag.Parse()

	if *brightnessFlag < -1 ||
		*brightnessFlag > 100 {
		fail(
			"brilho deve estar entre 0 e 100",
			2,
		)
	}

	cfg, err :=
		config.Load()

	if err != nil {
		fail(
			fmt.Sprintf(
				"carregar configuração: %v",
				err,
			),
			1,
		)
	}

	explicitScreen := false
	changed := false

	if *screenFlag != "" {
		screen, err :=
			r15m.ParseScreen(
				*screenFlag,
			)

		if err != nil {
			fail(
				err.Error(),
				2,
			)
		}

		cfg.LastScreen =
			screenConfigName(
				screen,
			)

		explicitScreen = true
		changed = true
	}

	if *brightnessFlag >= 0 {
		value :=
			*brightnessFlag

		cfg.Brightness =
			&value

		changed = true
	}

	if changed {
		if err := config.Save(
			cfg,
		); err != nil {
			fail(
				fmt.Sprintf(
					"salvar configuração: %v",
					err,
				),
				1,
			)
		}
	}

	printHeader(cfg)

	if *noRestore &&
		!explicitScreen {
		runWithoutRestore(cfg)
		return
	}

	if !cfg.RestoreLastScreen &&
		!explicitScreen {
		runWithoutRestore(cfg)
		return
	}

	screen, err :=
		r15m.ParseScreen(
			cfg.LastScreen,
		)

	if err != nil {
		fail(
			fmt.Sprintf(
				"interpretar última tela: %v",
				err,
			),
			1,
		)
	}

	ctx, stop :=
		signal.NotifyContext(
			context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)

	defer stop()

	interval :=
		time.Duration(
			cfg.MonitorIntervalSeconds,
		) * time.Second

	fmt.Printf(
		"Tela desejada : %s (%d)\n",
		screen.String(),
		screen,
	)

	fmt.Printf(
		"Atualização   : %s\n",
		interval,
	)

	fmt.Println(
		"Reconexão     : automática",
	)

	if !*once {
		fmt.Println(
			"Ctrl+C        : encerrar",
		)
	}

	fmt.Println()

	err =
		app.Run(
			ctx,
			app.Options{
				Screen: screen,

				Brightness: cfg.Brightness,

				City: cfg.City,

				WeatherAPIKey: cfg.WeatherAPIKey,

				MonitorInterval: interval,

				ReconnectDelay: app.DefaultReconnectDelay,

				Once: *once,

				Logf: func(
					format string,
					args ...any,
				) {
					fmt.Printf(
						format+"\n",
						args...,
					)
				},
			},
		)

	if err != nil {
		fail(
			fmt.Sprintf(
				"MiniTela: %v",
				err,
			),
			1,
		)
	}

	if !*once {
		fmt.Println()
		fmt.Println(
			"MiniTela encerrada.",
		)
	}
}

func printHeader(
	cfg config.Config,
) {
	fmt.Printf(
		"%s %s\n",
		appName,
		appVersion,
	)

	fmt.Println(
		"MiniTela para Positivo R15M",
	)

	fmt.Printf(
		"Sistema       : %s/%s\n",
		runtime.GOOS,
		runtime.GOARCH,
	)

	fmt.Println()

	path, _ :=
		config.Path()

	fmt.Printf(
		"Configuração  : %s\n",
		path,
	)

	fmt.Printf(
		"Última tela   : %s\n",
		cfg.LastScreen,
	)

	fmt.Printf(
		"Cidade        : %s\n",
		cfg.City,
	)

	if cfg.Brightness == nil {
		fmt.Println(
			"Brilho        : manter atual",
		)
	} else {
		fmt.Printf(
			"Brilho        : %d%%\n",
			*cfg.Brightness,
		)
	}

	if cfg.WeatherAPIKey == "" {
		fmt.Println(
			"WeatherAPI    : não configurada",
		)
	} else {
		fmt.Println(
			"WeatherAPI    : configurada",
		)
	}

	fmt.Printf(
		"Restaurar     : %t\n",
		cfg.RestoreLastScreen,
	)

	fmt.Println()
}

func runWithoutRestore(
	cfg config.Config,
) {
	fmt.Println(
		"Modo          : sem restauração",
	)

	fmt.Println(
		"Conectando somente para diagnóstico...",
	)

	conn, err :=
		r15m.Connect()

	if err != nil {
		fail(
			err.Error(),
			1,
		)
	}

	defer conn.Close()

	if cfg.Brightness != nil {
		if _, err :=
			conn.WriteRegisterVerified(
				r15m.RegisterBrightness,
				uint32(
					*cfg.Brightness,
				),
			); err != nil {
			fail(
				fmt.Sprintf(
					"aplicar brilho: %v",
					err,
				),
				1,
			)
		}
	}

	page, err :=
		conn.ReadRegister(
			r15m.RegisterCurrentPage,
		)

	if err != nil {
		fail(
			fmt.Sprintf(
				"ler página: %v",
				err,
			),
			1,
		)
	}

	fmt.Printf(
		"Tela atual    : %s (%d)\n",
		r15m.Screen(page).String(),
		page,
	)

	fmt.Println(
		"Status        : pronta",
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

func fail(
	message string,
	code int,
) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %s\n",
		message,
	)

	os.Exit(code)
}
