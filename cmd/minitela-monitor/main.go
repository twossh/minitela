package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twossh/minitela/internal/r15m"
)

const (
	reconnectDelay = 2 * time.Second
)

func main() {
	interval := flag.Duration(
		"interval",
		10*time.Second,
		"intervalo de atualização",
	)

	once := flag.Bool(
		"once",
		false,
		"executa somente uma atualização",
	)

	flag.Parse()

	if *interval < time.Second {
		fmt.Fprintln(
			os.Stderr,
			"intervalo mínimo: 1s",
		)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	fmt.Println("MiniTela Monitor")
	fmt.Println()

	if *once {
		if err := runOnce(ctx); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Erro: %v\n",
				err,
			)
			os.Exit(1)
		}

		return
	}

	fmt.Printf(
		"Atualização: %s\n",
		interval.String(),
	)

	fmt.Printf(
		"Reconexão automática: %s\n",
		reconnectDelay.String(),
	)

	fmt.Println(
		"Ctrl+C para encerrar.",
	)

	fmt.Println()

	runContinuous(
		ctx,
		*interval,
	)
}

func runOnce(
	ctx context.Context,
) error {
	conn, syncer, err := connectMonitor(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	result, err := syncer.Sync()
	if err != nil {
		return err
	}

	printSnapshot(result)

	return nil
}

func runContinuous(
	ctx context.Context,
	interval time.Duration,
) {
	for {
		if ctx.Err() != nil {
			fmt.Println()
			fmt.Println(
				"MiniTela Monitor encerrado.",
			)
			return
		}

		conn, syncer, err :=
			connectMonitor(ctx)

		if err != nil {
			if ctx.Err() != nil {
				fmt.Println()
				fmt.Println(
					"MiniTela Monitor encerrado.",
				)
				return
			}

			fmt.Printf(
				"[%s] aguardando MiniTela: %v\n",
				now(),
				err,
			)

			if !waitContext(
				ctx,
				reconnectDelay,
			) {
				return
			}

			continue
		}

		fmt.Printf(
			"[%s] conectado em %s\n",
			now(),
			conn.Device.Path,
		)

		err = monitorConnection(
			ctx,
			conn,
			syncer,
			interval,
		)

		_ = conn.Close()

		if ctx.Err() != nil {
			fmt.Println()
			fmt.Println(
				"MiniTela Monitor encerrado.",
			)
			return
		}

		if err != nil {
			fmt.Printf(
				"[%s] conexão perdida: %v\n",
				now(),
				err,
			)
		}

		fmt.Printf(
			"[%s] tentando reconectar...\n",
			now(),
		)

		if !waitContext(
			ctx,
			reconnectDelay,
		) {
			return
		}
	}
}

func connectMonitor(
	ctx context.Context,
) (
	*r15m.Connection,
	*r15m.MonitorSyncer,
	error,
) {
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	conn, err := r15m.Connect()
	if err != nil {
		return nil, nil, err
	}

	success := false

	defer func() {
		if !success {
			_ = conn.Close()
		}
	}()

	// Sempre restauramos a página Monitor depois de
	// uma nova conexão.
	if err := conn.SetScreen(
		r15m.ScreenMonitor,
	); err != nil {
		return nil, nil, fmt.Errorf(
			"restaurar tela Monitor: %w",
			err,
		)
	}

	// Novo sincronizador = cache vazio.
	// Isso força uma atualização completa após
	// reconectar ou voltar da suspensão.
	syncer := r15m.NewMonitorSyncer(
		conn,
	)

	success = true

	return conn, syncer, nil
}

func monitorConnection(
	ctx context.Context,
	conn *r15m.Connection,
	syncer *r15m.MonitorSyncer,
	interval time.Duration,
) error {
	// Primeiro ciclo imediatamente após conectar.
	if err := syncAndPrint(
		syncer,
	); err != nil {
		return err
	}

	ticker := time.NewTicker(
		interval,
	)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := syncAndPrint(
				syncer,
			); err != nil {
				return err
			}
		}
	}
}

func syncAndPrint(
	syncer *r15m.MonitorSyncer,
) error {
	result, err := syncer.Sync()
	if err != nil {
		return err
	}

	printSnapshot(result)

	return nil
}

func printSnapshot(
	result *r15m.MonitorSnapshot,
) {
	wifi := "desconectado"

	if result.WiFiConnected {
		wifi = fmt.Sprintf(
			"%s (%d%%)",
			result.WiFiDisplay,
			result.WiFiQuality,
		)
	}

	bluetooth := "desconectado"

	if result.BluetoothConnected {
		bluetooth =
			result.BluetoothDisplay
	}

	fmt.Printf(
		"[%s] bateria=%d%% nível=%d | wifi=%s | bluetooth=%s | escritas=%d\n",
		now(),
		result.BatteryPercent,
		result.BatteryLevel,
		wifi,
		bluetooth,
		result.Writes,
	)
}

func waitContext(
	ctx context.Context,
	duration time.Duration,
) bool {
	timer := time.NewTimer(
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
