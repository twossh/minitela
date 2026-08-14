package textureupload

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/twossh/minitela/internal/acf"
	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/r15m"
)

const (
	serviceName = "minitela.service"

	uploadTimeout = 10 * time.Minute

	controllerRestartDelay = 4 * time.Second
)

// Stage identifica as etapas observáveis do fluxo de upload.
type Stage string

const (
	StagePreparing       Stage = "preparing"
	StageConnecting      Stage = "connecting"
	StageUploading       Stage = "uploading"
	StageReconnecting    Stage = "reconnecting"
	StageSelectingScreen Stage = "selecting-screen"
	StageDone            Stage = "done"
)

func reportStage(
	fn func(Stage),
	stage Stage,
) {
	if fn != nil {
		fn(stage)
	}
}

// SendFile envia um ACF válido para o R15M.
//
// A função:
//
//   - valida checksum/footer;
//   - preserva o estado do minitela.service;
//   - faz o upload;
//   - aguarda a reenumeração USB;
//   - seleciona explicitamente a Página 5;
//   - salva "image" como última tela;
//   - restaura o serviço se ele estava ativo.
func SendFile(
	path string,
	onProgress func(r15m.UploadProgress),
) (
	result *r15m.UploadResult,
	err error,
) {
	return SendFileWithStage(
		path,
		onProgress,
		nil,
	)
}

// SendFileWithStage preserva o comportamento de SendFile e
// também reporta as transições de etapa para a interface.
func SendFileWithStage(
	path string,
	onProgress func(r15m.UploadProgress),
	onStage func(Stage),
) (
	result *r15m.UploadResult,
	err error,
) {
	reportStage(
		onStage,
		StagePreparing,
	)

	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf(
			"arquivo ACF não informado",
		)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"ler ACF %q: %w",
			path,
			err,
		)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf(
			"arquivo ACF vazio",
		)
	}

	if err := acf.ValidateChecksum(
		data,
	); err != nil {
		return nil, fmt.Errorf(
			"ACF inválido: %w",
			err,
		)
	}

	wasActive := serviceIsActive()

	if wasActive {
		if err := serviceAction(
			"stop",
		); err != nil {
			return nil, fmt.Errorf(
				"parar %s: %w",
				serviceName,
				err,
			)
		}

		defer func() {
			restoreErr :=
				serviceAction(
					"start",
				)

			if restoreErr != nil &&
				err == nil {

				err = fmt.Errorf(
					"upload concluído, mas falhou ao restaurar %s: %w",
					serviceName,
					restoreErr,
				)
			}
		}()
	}

	reportStage(
		onStage,
		StageConnecting,
	)

	conn, err := r15m.Connect()
	if err != nil {
		return nil, fmt.Errorf(
			"conectar ao R15M: %w",
			err,
		)
	}

	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			uploadTimeout,
		)

	defer cancel()

	reportStage(
		onStage,
		StageUploading,
	)

	result, err =
		conn.UploadTexture(
			ctx,
			data,
			onProgress,
		)

	_ = conn.Close()

	if err != nil {
		return nil, fmt.Errorf(
			"upload da textura: %w",
			err,
		)
	}

	reportStage(
		onStage,
		StageReconnecting,
	)

	// O DownloadComplete normalmente provoca reboot/
	// reenumeração do controlador USB.
	time.Sleep(
		controllerRestartDelay,
	)

	// Reconecta depois do reboot e seleciona
	// explicitamente a Página 5.
	pageConn, err :=
		r15m.Connect()

	if err != nil {
		return result, fmt.Errorf(
			"upload concluído, mas não foi possível reconectar ao R15M: %w",
			err,
		)
	}

	reportStage(
		onStage,
		StageSelectingScreen,
	)

	if err :=
		pageConn.SetScreen(
			r15m.ScreenImage,
		); err != nil {

		_ = pageConn.Close()

		return result, fmt.Errorf(
			"selecionar página Imagem: %w",
			err,
		)
	}

	_ = pageConn.Close()

	if err := saveImageAsLastScreen(); err != nil {

		return result, fmt.Errorf(
			"salvar Imagem como última tela: %w",
			err,
		)
	}

	reportStage(
		onStage,
		StageDone,
	)

	return result, nil
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
	cmd :=
		exec.Command(
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
	cmd :=
		exec.Command(
			"systemctl",
			"--user",
			action,
			serviceName,
		)

	output, err :=
		cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"systemctl %s: %w: %s",
			action,
			err,
			strings.TrimSpace(
				string(output),
			),
		)
	}

	return nil
}
