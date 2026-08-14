package main

import (
	"fmt"
	_ "image/gif"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/gallery"
)

const serviceName = "minitela.service"

func main() {
	a := app.NewWithID(
		"com.twossh.minitela",
	)

	w := a.NewWindow(
		"MiniTela",
	)

	w.Resize(
		fyne.NewSize(
			920,
			720,
		),
	)

	cfg, err := config.Load()
	if err != nil {
		dialog.ShowError(
			err,
			w,
		)

		return
	}

	serviceStatus :=
		widget.NewLabel("")

	actionStatus :=
		widget.NewLabel(
			"Pronto",
		)

	busy := false

	refreshStatus := func() {
		if serviceActive() {
			serviceStatus.SetText(
				"● MiniTela conectada / serviço ativo",
			)
		} else {
			serviceStatus.SetText(
				"○ Serviço MiniTela parado",
			)
		}
	}

	refreshStatus()

	runBackground := func(
		message string,
		success string,
		action func() error,
	) {
		if busy {
			dialog.ShowInformation(
				"MiniTela",
				"Há uma operação em andamento.",
				w,
			)

			return
		}

		busy = true

		actionStatus.SetText(
			message,
		)

		go func() {
			err := action()

			fyne.Do(
				func() {
					busy = false

					refreshStatus()

					if err != nil {
						actionStatus.SetText(
							"Erro",
						)

						dialog.ShowError(
							err,
							w,
						)

						return
					}

					actionStatus.SetText(
						success,
					)
				},
			)
		}()
	}

	//
	// CONTROLE DE ENERGIA
	//

	onButton :=
		widget.NewButton(
			"Ligar",
			func() {
				runBackground(
					"Ligando MiniTela...",
					"MiniTela ligada.",
					func() error {
						return runSibling(
							"MiniTelaCtl",
							"--on",
						)
					},
				)
			},
		)

	offButton :=
		widget.NewButton(
			"Desligar",
			func() {
				runBackground(
					"Desligando MiniTela...",
					"MiniTela desligada.",
					func() error {
						return runSibling(
							"MiniTelaCtl",
							"--off",
						)
					},
				)
			},
		)

	rebootButton :=
		widget.NewButton(
			"Reiniciar",
			func() {
				dialog.ShowConfirm(
					"Reiniciar MiniTela",
					"Reiniciar o controlador da MiniTela?",
					func(ok bool) {
						if !ok {
							return
						}

						runBackground(
							"Reiniciando controlador...",
							"MiniTela reiniciada.",
							func() error {
								return runSibling(
									"MiniTelaCtl",
									"--reboot",
								)
							},
						)
					},
					w,
				)
			},
		)

	restartServiceButton :=
		widget.NewButton(
			"Recarregar",
			func() {
				runBackground(
					"Recarregando configurações...",
					"Configurações aplicadas.",
					func() error {
						return runSibling(
							"MiniTelaCtl",
							"--restart",
						)
					},
				)
			},
		)

	powerButtons :=
		container.NewGridWithColumns(
			4,
			onButton,
			offButton,
			rebootButton,
			restartServiceButton,
		)

	//
	// TELA
	//

	screenSelect :=
		widget.NewSelect(
			[]string{
				"Monitor",
				"Clima",
				"Notas",
				"WhatsApp",
				"Imagem",
			},
			nil,
		)

	screenSelect.SetSelected(
		screenDisplayName(
			cfg.LastScreen,
		),
	)

	//
	// CIDADE
	//

	cityEntry :=
		widget.NewEntry()

	cityEntry.SetPlaceHolder(
		"Ex.: Porto Alegre",
	)

	cityEntry.SetText(
		cfg.City,
	)

	//
	// BRILHO
	//

	brightnessValue := 50

	if cfg.Brightness != nil {
		brightnessValue =
			*cfg.Brightness
	}

	brightnessLabel :=
		widget.NewLabel(
			fmt.Sprintf(
				"%d%%",
				brightnessValue,
			),
		)

	brightnessSlider :=
		widget.NewSlider(
			0,
			100,
		)

	brightnessSlider.Step = 1

	brightnessSlider.SetValue(
		float64(
			brightnessValue,
		),
	)

	brightnessSlider.OnChanged =
		func(value float64) {
			brightnessLabel.SetText(
				fmt.Sprintf(
					"%.0f%%",
					value,
				),
			)
		}

	brightnessRow :=
		container.NewBorder(
			nil,
			nil,
			nil,
			brightnessLabel,
			brightnessSlider,
		)

	//
	// OPÇÕES
	//

	restoreCheck :=
		widget.NewCheck(
			"Restaurar última tela ao iniciar",
			nil,
		)

	restoreCheck.SetChecked(
		cfg.RestoreLastScreen,
	)

	autostartCheck :=
		widget.NewCheck(
			"Iniciar com o sistema",
			nil,
		)

	autostartCheck.SetChecked(
		serviceEnabled(),
	)

	//
	// APLICAR CONFIGURAÇÕES
	//

	applyButton :=
		widget.NewButton(
			"Aplicar configurações",
			func() {
				runBackground(
					"Aplicando configurações...",
					"Configurações aplicadas.",
					func() error {
						current, err :=
							config.Load()

						if err != nil {
							return err
						}

						current.LastScreen =
							screenConfigName(
								screenSelect.Selected,
							)

						current.City =
							strings.TrimSpace(
								cityEntry.Text,
							)

						brightness :=
							int(
								brightnessSlider.Value +
									0.5,
							)

						current.Brightness =
							&brightness

						current.RestoreLastScreen =
							restoreCheck.Checked

						current.Autostart =
							autostartCheck.Checked

						if err :=
							config.Save(
								current,
							); err != nil {
							return err
						}

						if autostartCheck.Checked {
							if err :=
								systemctl(
									"enable",
								); err != nil {
								return err
							}
						} else {
							if err :=
								systemctl(
									"disable",
								); err != nil {
								return err
							}
						}

						return systemctl(
							"restart",
						)
					},
				)
			},
		)

	//
	// PAINEL PRINCIPAL
	//

	title :=
		widget.NewLabelWithStyle(
			"MiniTela",
			fyne.TextAlignLeading,
			fyne.TextStyle{
				Bold: true,
			},
		)

	subtitle :=
		widget.NewLabel(
			"Positivo Vision R15M",
		)

	controlContent :=
		container.NewVBox(
			title,
			subtitle,

			widget.NewSeparator(),

			serviceStatus,
			actionStatus,

			widget.NewSeparator(),

			widget.NewLabelWithStyle(
				"Controle",
				fyne.TextAlignLeading,
				fyne.TextStyle{
					Bold: true,
				},
			),

			powerButtons,

			widget.NewSeparator(),

			widget.NewLabelWithStyle(
				"Tela",
				fyne.TextAlignLeading,
				fyne.TextStyle{
					Bold: true,
				},
			),

			screenSelect,

			widget.NewLabelWithStyle(
				"Cidade do clima",
				fyne.TextAlignLeading,
				fyne.TextStyle{
					Bold: true,
				},
			),

			cityEntry,

			widget.NewLabelWithStyle(
				"Brilho",
				fyne.TextAlignLeading,
				fyne.TextStyle{
					Bold: true,
				},
			),

			brightnessRow,

			widget.NewSeparator(),

			restoreCheck,
			autostartCheck,

			applyButton,
		)

	controlScroll :=
		container.NewVScroll(
			controlContent,
		)

	//
	// GALERIA ORIGINAL
	//

	galleryContent :=
		buildGallery(
			w,
			runBackground,
		)

	galleryScroll :=
		container.NewVScroll(
			galleryContent,
		)

	//
	// ABAS
	//

	tabs :=
		container.NewAppTabs(
			container.NewTabItem(
				"Controle",
				controlScroll,
			),

			container.NewTabItem(
				"Galeria",
				galleryScroll,
			),
		)

	tabs.SetTabLocation(
		container.TabLocationTop,
	)

	w.SetContent(
		container.NewPadded(
			tabs,
		),
	)

	w.ShowAndRun()
}

func buildGallery(
	w fyne.Window,
	runBackground func(
		string,
		string,
		func() error,
	),
) fyne.CanvasObject {
	items, err :=
		gallery.Load()

	if err != nil {
		return container.NewVBox(
			widget.NewLabel(
				"Galeria original não encontrada.",
			),

			widget.NewLabel(
				err.Error(),
			),
		)
	}

	objects :=
		make(
			[]fyne.CanvasObject,
			0,
			len(items),
		)

	for _, item := range items {

		item := item
		id := item.ID

		preview :=
			canvas.NewImageFromFile(
				item.PreviewPath,
			)

		preview.FillMode =
			canvas.ImageFillContain

		preview.SetMinSize(
			fyne.NewSize(
				150,
				150,
			),
		)

		send :=
			widget.NewButton(
				"Enviar",
				func() {
					runBackground(
						fmt.Sprintf(
							"Enviando imagem %d...",
							id,
						),

						fmt.Sprintf(
							"Imagem %d enviada.",
							id,
						),

						func() error {
							return runSibling(
								"MiniTelaImages",
								"--send",
								strconv.Itoa(id),
							)
						},
					)
				},
			)

		cardBody :=
			container.NewBorder(
				nil,
				send,
				nil,
				nil,
				preview,
			)

		card :=
			widget.NewCard(
				fmt.Sprintf(
					"Imagem %02d",
					id,
				),

				formatSize(
					item.TextureSize,
				),

				cardBody,
			)

		objects =
			append(
				objects,
				card,
			)
	}

	grid :=
		container.New(
			layout.NewGridLayout(
				3,
			),
			objects...,
		)

	return container.NewVBox(
		widget.NewLabelWithStyle(
			"Galeria Original",
			fyne.TextAlignLeading,
			fyne.TextStyle{
				Bold: true,
			},
		),

		widget.NewLabel(
			"Selecione uma das 21 animações e envie para a MiniTela.",
		),

		widget.NewSeparator(),

		grid,

		widget.NewSeparator(),

		widget.NewButton(
			"Selecionar imagem própria — próxima etapa",
			func() {
				dialog.ShowInformation(
					"Imagem própria",
					"O suporte nativo para PNG, JPG e GIF será adicionado na próxima etapa.",
					w,
				)
			},
		),
	)
}

func siblingBinary(
	name string,
) string {
	executable, err :=
		os.Executable()

	if err == nil {
		candidate :=
			filepath.Join(
				filepath.Dir(
					executable,
				),
				name,
			)

		if _, err :=
			os.Stat(
				candidate,
			); err == nil {
			return candidate
		}
	}

	return name
}

func runSibling(
	name string,
	args ...string,
) error {
	command :=
		exec.Command(
			siblingBinary(name),
			args...,
		)

	output, err :=
		command.CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"%s: %w\n%s",
			name,
			err,
			strings.TrimSpace(
				string(output),
			),
		)
	}

	return nil
}

func systemctl(
	action string,
) error {
	command :=
		exec.Command(
			"systemctl",
			"--user",
			action,
			serviceName,
		)

	output, err :=
		command.CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"systemctl %s: %w\n%s",
			action,
			err,
			strings.TrimSpace(
				string(output),
			),
		)
	}

	return nil
}

func serviceActive() bool {
	command :=
		exec.Command(
			"systemctl",
			"--user",
			"is-active",
			"--quiet",
			serviceName,
		)

	return command.Run() == nil
}

func serviceEnabled() bool {
	command :=
		exec.Command(
			"systemctl",
			"--user",
			"is-enabled",
			"--quiet",
			serviceName,
		)

	return command.Run() == nil
}

func screenDisplayName(
	value string,
) string {
	switch value {
	case "whatsapp":
		return "WhatsApp"

	case "notes":
		return "Notas"

	case "weather":
		return "Clima"

	case "image":
		return "Imagem"

	default:
		return "Monitor"
	}
}

func screenConfigName(
	value string,
) string {
	switch value {
	case "WhatsApp":
		return "whatsapp"

	case "Notas":
		return "notes"

	case "Clima":
		return "weather"

	case "Imagem":
		return "image"

	default:
		return "monitor"
	}
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
