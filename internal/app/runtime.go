package app

import (
	"context"
	"fmt"
	"time"

	"github.com/twossh/minitela/internal/r15m"
)

const (
	DefaultReconnectDelay = 2 * time.Second

	WeatherRefreshInterval = 30 * time.Minute

	WeatherRetryInterval = 5 * time.Minute
)

type Logger func(
	format string,
	args ...any,
)

type Options struct {
	Screen r15m.Screen

	Brightness *int

	City string

	WeatherAPIKey string

	MonitorInterval time.Duration

	ReconnectDelay time.Duration

	Once bool

	Logf Logger
}

func Run(
	ctx context.Context,
	opts Options,
) error {
	if opts.MonitorInterval <
		time.Second {
		opts.MonitorInterval =
			10 * time.Second
	}

	if opts.ReconnectDelay <
		time.Second {
		opts.ReconnectDelay =
			DefaultReconnectDelay
	}

	if opts.Logf == nil {
		opts.Logf =
			func(
				string,
				...any,
			) {
			}
	}

	if opts.Once {
		return runOnce(
			ctx,
			opts,
		)
	}

	return runContinuous(
		ctx,
		opts,
	)
}

func runOnce(
	ctx context.Context,
	opts Options,
) error {
	conn, err :=
		connectAndRestore(
			ctx,
			opts,
		)

	if err != nil {
		return err
	}

	defer conn.Close()

	switch opts.Screen {
	case r15m.ScreenMonitor:
		syncer :=
			r15m.NewMonitorSyncer(
				conn,
			)

		result, err :=
			syncer.Sync()

		if err != nil {
			return err
		}

		logMonitorSnapshot(
			opts.Logf,
			result,
		)

	case r15m.ScreenWeather:
		return syncWeatherOnce(
			ctx,
			conn,
			opts,
		)
	}

	return nil
}

func runContinuous(
	ctx context.Context,
	opts Options,
) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		conn, err :=
			connectAndRestore(
				ctx,
				opts,
			)

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			opts.Logf(
				"[%s] aguardando MiniTela: %v",
				now(),
				err,
			)

			if !waitContext(
				ctx,
				opts.ReconnectDelay,
			) {
				return nil
			}

			continue
		}

		opts.Logf(
			"[%s] conectado em %s",
			now(),
			conn.Device.Path,
		)

		err =
			runConnected(
				ctx,
				conn,
				opts,
			)

		_ = conn.Close()

		if ctx.Err() != nil {
			return nil
		}

		if err != nil {
			opts.Logf(
				"[%s] conexão perdida: %v",
				now(),
				err,
			)
		}

		opts.Logf(
			"[%s] tentando reconectar...",
			now(),
		)

		if !waitContext(
			ctx,
			opts.ReconnectDelay,
		) {
			return nil
		}
	}
}

func connectAndRestore(
	ctx context.Context,
	opts Options,
) (*r15m.Connection, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	conn, err :=
		r15m.Connect()

	if err != nil {
		return nil, err
	}

	success := false

	defer func() {
		if !success {
			_ = conn.Close()
		}
	}()

	if opts.Brightness != nil {
		value :=
			*opts.Brightness

		if value < 0 {
			value = 0
		}

		if value > 100 {
			value = 100
		}

		actual, err :=
			conn.WriteRegisterVerified(
				r15m.RegisterBrightness,
				uint32(value),
			)

		if err != nil {
			return nil, fmt.Errorf(
				"restaurar brilho: %w",
				err,
			)
		}

		opts.Logf(
			"[%s] brilho restaurado: %d%%",
			now(),
			actual,
		)
	}

	if err := conn.SetScreen(
		opts.Screen,
	); err != nil {
		return nil, fmt.Errorf(
			"restaurar tela %s: %w",
			opts.Screen.String(),
			err,
		)
	}

	opts.Logf(
		"[%s] tela restaurada: %s (%d)",
		now(),
		opts.Screen.String(),
		opts.Screen,
	)

	success = true

	return conn, nil
}

func runConnected(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	switch opts.Screen {
	case r15m.ScreenMonitor:
		return runMonitor(
			ctx,
			conn,
			opts,
		)

	case r15m.ScreenWeather:
		return runWeather(
			ctx,
			conn,
			opts,
		)

	default:
		return runStaticScreen(
			ctx,
			conn,
			opts,
		)
	}
}

func runMonitor(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	syncer :=
		r15m.NewMonitorSyncer(
			conn,
		)

	if err := syncMonitor(
		syncer,
		opts.Logf,
	); err != nil {
		return err
	}

	ticker :=
		time.NewTicker(
			opts.MonitorInterval,
		)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := syncMonitor(
				syncer,
				opts.Logf,
			); err != nil {
				return err
			}
		}
	}
}

func syncMonitor(
	syncer *r15m.MonitorSyncer,
	logf Logger,
) error {
	result, err :=
		syncer.Sync()

	if err != nil {
		return err
	}

	logMonitorSnapshot(
		logf,
		result,
	)

	return nil
}

func logMonitorSnapshot(
	logf Logger,
	result *r15m.MonitorSnapshot,
) {
	wifi := "desconectado"

	if result.WiFiConnected {
		wifi =
			fmt.Sprintf(
				"%s (%d%%)",
				result.WiFiDisplay,
				result.WiFiQuality,
			)
	}

	bluetooth :=
		"desconectado"

	if result.BluetoothConnected {
		bluetooth =
			result.BluetoothDisplay
	}

	logf(
		"[%s] bateria=%d%% nível=%d | wifi=%s | bluetooth=%s | escritas=%d",
		now(),
		result.BatteryPercent,
		result.BatteryLevel,
		wifi,
		bluetooth,
		result.Writes,
	)
}

func syncWeatherOnce(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	if opts.City == "" {
		return fmt.Errorf(
			"cidade não configurada",
		)
	}

	if opts.WeatherAPIKey == "" {
		return fmt.Errorf(
			"chave WeatherAPI não configurada",
		)
	}

	opts.Logf(
		"[%s] consultando clima para %s...",
		now(),
		opts.City,
	)

	syncer :=
		r15m.NewWeatherSyncer(
			conn,
			opts.City,
			opts.WeatherAPIKey,
		)

	result, err :=
		syncer.Sync(ctx)

	if err != nil {
		return fmt.Errorf(
			"consultar/sincronizar clima: %w",
			err,
		)
	}

	logWeatherSnapshot(
		opts.Logf,
		result,
	)

	return nil
}

func runWeather(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	if opts.City == "" {
		return fmt.Errorf(
			"cidade não configurada",
		)
	}

	if opts.WeatherAPIKey == "" {
		return fmt.Errorf(
			"chave WeatherAPI não configurada",
		)
	}

	syncer :=
		r15m.NewWeatherSyncer(
			conn,
			opts.City,
			opts.WeatherAPIKey,
		)

	opts.Logf(
		"[%s] consultando clima para %s...",
		now(),
		opts.City,
	)

	refreshDelay :=
		WeatherRefreshInterval

	result, err :=
		syncer.Sync(ctx)

	if err != nil {
		opts.Logf(
			"[%s] clima: consulta inicial falhou: %v",
			now(),
			err,
		)

		refreshDelay =
			WeatherRetryInterval
	} else {
		logWeatherSnapshot(
			opts.Logf,
			result,
		)
	}

	heartbeat :=
		time.NewTicker(
			opts.MonitorInterval,
		)

	defer heartbeat.Stop()

	refresh :=
		time.NewTimer(
			refreshDelay,
		)

	defer refresh.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-heartbeat.C:
			page, err :=
				conn.ReadRegister(
					r15m.RegisterCurrentPage,
				)

			if err != nil {
				return err
			}

			if page != uint32(
				r15m.ScreenWeather,
			) {
				if err :=
					conn.SetScreen(
						r15m.ScreenWeather,
					); err != nil {
					return err
				}
			}

		case <-refresh.C:
			opts.Logf(
				"[%s] atualizando clima...",
				now(),
			)

			result, err :=
				syncer.Sync(ctx)

			if err != nil {
				opts.Logf(
					"[%s] clima: atualização falhou: %v",
					now(),
					err,
				)

				refresh.Reset(
					WeatherRetryInterval,
				)

				continue
			}

			logWeatherSnapshot(
				opts.Logf,
				result,
			)

			refresh.Reset(
				WeatherRefreshInterval,
			)
		}
	}
}

func logWeatherSnapshot(
	logf Logger,
	result *r15m.WeatherSnapshot,
) {
	logf(
		"[%s] clima=%s | hoje=%s icon=%d | amanhã=%s %s icon=%d | terceiro=%s %s icon=%d | escritas=%d",
		now(),
		result.City,
		result.TodayTemp,
		result.TodayIcon,
		result.TomorrowDate,
		result.TomorrowTemp,
		result.TomorrowIcon,
		result.ThirdDate,
		result.ThirdTemp,
		result.ThirdIcon,
		result.Writes,
	)
}

func runStaticScreen(
	ctx context.Context,
	conn *r15m.Connection,
	opts Options,
) error {
	ticker :=
		time.NewTicker(
			opts.MonitorInterval,
		)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			page, err :=
				conn.ReadRegister(
					r15m.RegisterCurrentPage,
				)

			if err != nil {
				return err
			}

			if page != uint32(
				opts.Screen,
			) {
				opts.Logf(
					"[%s] página alterada (%d); restaurando %s...",
					now(),
					page,
					opts.Screen.String(),
				)

				if err :=
					conn.SetScreen(
						opts.Screen,
					); err != nil {
					return err
				}
			}
		}
	}
}

func waitContext(
	ctx context.Context,
	duration time.Duration,
) bool {
	timer :=
		time.NewTimer(
			duration,
		)

	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false

	case <-timer.C:
		return true
	}
}

func now() string {
	return time.Now().Format(
		"15:04:05",
	)
}
