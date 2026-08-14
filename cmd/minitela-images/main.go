package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/gallery"
	"github.com/twossh/minitela/internal/r15m"
)

const serviceName = "minitela.service"

func main() {
	list := flag.Bool(
		"list",
		false,
		"lista as imagens disponíveis",
	)

	info := flag.Int(
		"info",
		0,
		"exibe informações da imagem",
	)

	dryRun := flag.Int(
		"dry-run",
		0,
		"simula o upload sem acessar o R15M",
	)

	send := flag.Int(
		"send",
		0,
		"envia uma imagem da galeria para o R15M",
	)

	flag.Parse()

	selected := 0

	if *list {
		selected++
	}

	if *info != 0 {
		selected++
	}

	if *dryRun != 0 {
		selected++
	}

	if *send != 0 {
		selected++
	}

	if selected == 0 {
		showUsage()
		return
	}

	if selected != 1 {
		fail(fmt.Errorf(
			"use somente uma opção por vez",
		))
	}

	items, err :=
		gallery.Load()

	if err != nil {
		fail(err)
	}

	switch {
	case *list:
		showList(items)

	case *info != 0:
		item, err :=
			gallery.Find(
				items,
				*info,
			)

		if err != nil {
			fail(err)
		}

		showItem(item)

	case *dryRun != 0:
		item, err :=
			gallery.Find(
				items,
				*dryRun,
			)

		if err != nil {
			fail(err)
		}

		plan, err :=
			gallery.BuildUploadPlan(
				item,
			)

		if err != nil {
			fail(err)
		}

		showDryRun(plan)

	case *send != 0:
		item, err :=
			gallery.Find(
				items,
				*send,
			)

		if err != nil {
			fail(err)
		}

		if err :=
			sendTexture(item); err != nil {
			fail(err)
		}
	}
}

func sendTexture(
	item gallery.Item,
) error {
	if err :=
		gallery.ValidateTexture(
			item,
		); err != nil {
		return err
	}

	data, err :=
		os.ReadFile(
			item.TexturePath,
		)

	if err != nil {
		return err
	}

	plan, err :=
		gallery.BuildUploadPlan(
			item,
		)

	if err != nil {
		return err
	}

	fmt.Println(
		"=== MiniTela - Envio de Imagem ===",
	)

	fmt.Println()

	fmt.Printf(
		"Imagem        : %d\n",
		item.ID,
	)

	fmt.Printf(
		"Arquivo       : %s\n",
		item.TexturePath,
	)

	fmt.Printf(
		"Tamanho       : %s\n",
		formatSize(
			item.TextureSize,
		),
	)

	fmt.Printf(
		"MD5/File ID   : %x\n",
		plan.FileID,
	)

	fmt.Printf(
		"Destino       : 0x%08X\n",
		gallery.TextureAddress,
	)

	fmt.Println()

	wasActive :=
		serviceIsActive()

	if wasActive {
		fmt.Println(
			"Parando minitela.service...",
		)

		if err :=
			serviceAction(
				"stop",
			); err != nil {
			return fmt.Errorf(
				"parar serviço: %w",
				err,
			)
		}
	}

	restoreService := wasActive

	defer func() {
		if restoreService {
			fmt.Println()
			fmt.Println(
				"Restaurando minitela.service...",
			)

			if err :=
				serviceAction(
					"start",
				); err != nil {

				fmt.Fprintf(
					os.Stderr,
					"AVISO: falha ao restaurar serviço: %v\n",
					err,
				)

				return
			}

			fmt.Println(
				"Serviço restaurado.",
			)
		}
	}()

	fmt.Println(
		"Conectando ao R15M...",
	)

	conn, err :=
		r15m.Connect()

	if err != nil {
		return fmt.Errorf(
			"conectar R15M: %w",
			err,
		)
	}

	fmt.Printf(
		"Dispositivo   : %s\n",
		conn.Device.Path,
	)

	fmt.Println(
		"Handshake     : OK",
	)

	fmt.Println()
	fmt.Println(
		"Iniciando upload...",
	)

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			10*time.Minute,
		)

	defer cancel()

	lastPrinted := -1

	result, err :=
		conn.UploadTexture(
			ctx,
			data,
			func(
				progress r15m.UploadProgress,
			) {
				displayPercent :=
					(progress.Percent / 5) * 5

				if displayPercent >
					100 {
					displayPercent = 100
				}

				if displayPercent ==
					lastPrinted {
					return
				}

				lastPrinted =
					displayPercent

				fmt.Printf(
					"Progresso     : %3d%% (%d/%d bytes)\n",
					displayPercent,
					progress.BytesSent,
					progress.TotalBytes,
				)
			},
		)

	_ = conn.Close()

	if err != nil {
		return fmt.Errorf(
			"upload: %w",
			err,
		)
	}

	if result.AlreadyPresent {
		fmt.Println()
		fmt.Println(
			"A textura já estava instalada no R15M.",
		)
	} else {
		fmt.Println()
		fmt.Printf(
			"MaxPageSize   : %d bytes\n",
			result.MaxPageSize,
		)

		fmt.Printf(
			"Offset inicial: %d bytes\n",
			result.StartOffset,
		)

		fmt.Println(
			"Transferência : concluída",
		)
	}

	// O firmware pode reenumerar a USB no fim
	// do DownloadComplete.
	fmt.Println()
	fmt.Println(
		"Aguardando o controlador...",
	)

	time.Sleep(
		4 * time.Second,
	)

	fmt.Println(
		"Upload        : OK",
	)

	if err := saveImageAsLastScreen(); err != nil {
		return fmt.Errorf(
			"salvar imagem como última tela: %w",
			err,
		)
	}

	fmt.Println(
		"Última tela   : Imagem (5)",
	)

	return nil
}

func saveImageAsLastScreen() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cfg.LastScreen = "image"
	cfg.RestoreLastScreen = true

	return config.Save(cfg)
}

func serviceIsActive() bool {
	cmd := exec.Command(
		"systemctl",
		"--user",
		"is-active",
		"--quiet",
		serviceName,
	)

	return cmd.Run() == nil
}

func serviceAction(
	action string,
) error {
	cmd := exec.Command(
		"systemctl",
		"--user",
		action,
		serviceName,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func showUsage() {
	fmt.Println(
		"MiniTela Images",
	)

	fmt.Println()
	fmt.Println(
		"Uso:",
	)

	fmt.Println(
		"  MiniTelaImages --list",
	)

	fmt.Println(
		"  MiniTelaImages --info 1",
	)

	fmt.Println(
		"  MiniTelaImages --dry-run 1",
	)

	fmt.Println(
		"  MiniTelaImages --send 1",
	)
}

func showList(
	items []gallery.Item,
) {
	fmt.Println(
		"=== MiniTela - Galeria Original ===",
	)

	fmt.Println()

	fmt.Printf(
		"Imagens disponíveis: %d\n",
		len(items),
	)

	fmt.Println()

	fmt.Printf(
		"%-5s %-12s %-12s\n",
		"ID",
		"GIF",
		"ACF",
	)

	fmt.Printf(
		"%-5s %-12s %-12s\n",
		"--",
		"---",
		"---",
	)

	for _, item := range items {
		fmt.Printf(
			"%-5d %-12s %-12s\n",
			item.ID,
			formatSize(
				item.PreviewSize,
			),
			formatSize(
				item.TextureSize,
			),
		)
	}
}

func showItem(
	item gallery.Item,
) {
	fmt.Printf(
		"Imagem       : %d\n",
		item.ID,
	)

	fmt.Printf(
		"Preview      : %s\n",
		item.PreviewPath,
	)

	fmt.Printf(
		"Preview size : %s\n",
		formatSize(
			item.PreviewSize,
		),
	)

	fmt.Printf(
		"Texture      : %s\n",
		item.TexturePath,
	)

	fmt.Printf(
		"Texture size : %s\n",
		formatSize(
			item.TextureSize,
		),
	)
}

func showDryRun(
	plan *gallery.UploadPlan,
) {
	fmt.Println(
		"=== MiniTela - Dry Run de Upload ===",
	)

	fmt.Println()
	fmt.Println(
		"NENHUM DADO SERÁ ENVIADO AO R15M.",
	)

	fmt.Println()

	fmt.Printf(
		"Imagem           : %d\n",
		plan.Item.ID,
	)

	fmt.Printf(
		"Arquivo          : %s\n",
		plan.Item.TexturePath,
	)

	fmt.Printf(
		"Tamanho          : %s (%d bytes)\n",
		formatSize(
			plan.FileSize,
		),
		plan.FileSize,
	)

	fmt.Printf(
		"MD5/File ID      : %x\n",
		plan.FileID,
	)

	fmt.Println(
		"Destino          : texture",
	)

	fmt.Printf(
		"Endereço R15M    : 0x%08X\n",
		plan.Address,
	)

	fmt.Println()

	fmt.Printf(
		"PageSize simulado: %d bytes\n",
		plan.SimulatedPageSize,
	)

	fmt.Printf(
		"Chunks simulados : %d\n",
		plan.SimulatedChunks,
	)

	fmt.Printf(
		"Último chunk     : %d bytes\n",
		plan.LastChunkSize,
	)

	fmt.Println()

	fmt.Println(
		"Validação ACF    : OK",
	)

	fmt.Println(
		"Dry-run          : OK",
	)
}

func formatSize(
	bytes int64,
) string {
	const (
		kb int64 = 1024
		mb       = 1024 * kb
	)

	switch {
	case bytes >= mb:
		return fmt.Sprintf(
			"%.1f MB",
			float64(bytes)/
				float64(mb),
		)

	case bytes >= kb:
		return fmt.Sprintf(
			"%.1f KB",
			float64(bytes)/
				float64(kb),
		)

	default:
		return fmt.Sprintf(
			"%d B",
			bytes,
		)
	}
}

func fail(err error) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %v\n",
		err,
	)

	os.Exit(1)
}
