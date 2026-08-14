package main

import (
	"fmt"
	_ "image/gif"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/twossh/minitela/internal/config"
	"github.com/twossh/minitela/internal/customimage"
	"github.com/twossh/minitela/internal/device"
	"github.com/twossh/minitela/internal/gallery"
	"github.com/twossh/minitela/internal/r15m"
	"github.com/twossh/minitela/internal/textureupload"
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
			900,
			580,
		),
	)

	w.CenterOnScreen()

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

	environmentStatus :=
		widget.NewLabel("")
	environmentStatus.Wrapping =
		fyne.TextWrapWord

	actionStatus :=
		widget.NewLabel(
			"Pronto",
		)

	busy := false

	refreshStatus := func() {
		check := inspectEnvironment()

		serviceStatus.SetText(
			check.primaryStatus(),
		)

		environmentStatus.SetText(
			check.detailStatus(),
		)
	}

	refreshStatus()

	go func() {
		ticker := time.NewTicker(
			2 * time.Second,
		)
		defer ticker.Stop()

		for range ticker.C {
			fyne.Do(
				func() {
					refreshStatus()
				},
			)
		}
	}()

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

	//
	// LAYOUT DA ABA CONTROLE
	//

	controlHeader :=
		container.NewVBox(
			title,
			subtitle,
			widget.NewSeparator(),
		)

	powerTop :=
		container.NewGridWithColumns(
			2,
			onButton,
			offButton,
		)

	powerGrid :=
		container.NewVBox(
			powerTop,
			rebootButton,
		)

	diagnosticButton :=
		widget.NewButton(
			"Diagnóstico",
			func() {
				dialog.ShowInformation(
					"Diagnóstico MiniTela",
					inspectEnvironment().report(),
					w,
				)
			},
		)

	controlCard :=
		widget.NewCard(
			"Controle",
			"Energia e estado da MiniTela",
			container.NewVBox(
				powerGrid,

				widget.NewSeparator(),

				serviceStatus,
				environmentStatus,
				actionStatus,

				widget.NewSeparator(),
				diagnosticButton,
			),
		)

	displayCard :=
		widget.NewCard(
			"Exibição",
			"Tela, cidade e preferências",
			container.NewVBox(
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

				container.NewGridWithColumns(
					2,
					restoreCheck,
					autostartCheck,
				),
			),
		)

	controlColumns :=
		container.NewGridWithColumns(
			2,
			controlCard,
			displayCard,
		)

	applyArea :=
		container.NewVBox(
			widget.NewSeparator(),
			applyButton,
		)

	controlScroll :=
		container.NewBorder(
			controlHeader,
			applyArea,
			nil,
			nil,
			controlColumns,
		)

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
	items, err := gallery.Load()

	if err != nil || len(items) == 0 {
		return buildCustomOnlyGallery(
			w,
			runBackground,
		)
	}

	selected := items[0]
	customSelected := false
	var customInfo *customimage.Info

	preview :=
		canvas.NewImageFromFile(
			selected.PreviewPath,
		)
	preview.FillMode =
		canvas.ImageFillContain
	preview.CornerRadius = 12
	preview.SetMinSize(
		fyne.NewSize(
			220,
			220,
		),
	)

	selectedTitle :=
		widget.NewLabelWithStyle(
			fmt.Sprintf(
				"Imagem %02d",
				selected.ID,
			),
			fyne.TextAlignCenter,
			fyne.TextStyle{
				Bold: true,
			},
		)

	selectedInfo :=
		widget.NewLabel(
			fmt.Sprintf(
				"ACF: %s",
				formatSize(
					selected.TextureSize,
				),
			),
		)
	selectedInfo.Alignment =
		fyne.TextAlignCenter

	sendButton :=
		widget.NewButton(
			"Enviar para MiniTela",
			nil,
		)
	sendButton.Importance =
		widget.HighImportance

	progressLabel :=
		widget.NewLabel("")
	progressLabel.Alignment =
		fyne.TextAlignCenter

	progressBar :=
		widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 1
	progressBar.SetValue(0)
	progressLabel.Hide()
	progressBar.Hide()

	var customButton *widget.Button
	galleryButtons := make(
		[]*widget.Button,
		0,
		len(items),
	)

	setGalleryBusy := func(value bool) {
		if value {
			sendButton.Disable()

			if customButton != nil {
				customButton.Disable()
			}

			for _, button := range galleryButtons {
				button.Disable()
			}

			return
		}

		sendButton.Enable()

		if customButton != nil {
			customButton.Enable()
		}

		for _, button := range galleryButtons {
			button.Enable()
		}
	}

	resetProgress := func() {
		progressBar.SetValue(0)
		progressLabel.Hide()
		progressBar.Hide()
	}

	showProgress := func(text string) {
		progressLabel.SetText(text)
		progressLabel.Show()
		progressBar.Show()
	}

	updateBuildStage := func(
		stage customimage.BuildStage,
	) {
		fyne.Do(
			func() {
				switch stage {
				case customimage.BuildStageImage:
					showProgress(
						"Gerando imagem...",
					)

				case customimage.BuildStageSTCRGBA:
					showProgress(
						"Convertendo STCRGBA...",
					)

				case customimage.BuildStageACF:
					showProgress(
						"Preparando ACF...",
					)
				}
			},
		)
	}

	updateUploadStage := func(
		stage textureupload.Stage,
	) {
		fyne.Do(
			func() {
				switch stage {
				case textureupload.StagePreparing:
					showProgress(
						"Preparando upload...",
					)

				case textureupload.StageConnecting:
					showProgress(
						"Conectando à MiniTela...",
					)

				case textureupload.StageUploading:
					progressBar.SetValue(0)
					showProgress(
						"Enviando 0%",
					)

				case textureupload.StageReconnecting:
					progressBar.SetValue(1)
					showProgress(
						"Reconectando MiniTela...",
					)

				case textureupload.StageSelectingScreen:
					progressBar.SetValue(1)
					showProgress(
						"Ativando imagem na MiniTela...",
					)

				case textureupload.StageDone:
					progressBar.SetValue(1)
					showProgress(
						"Imagem enviada ✓",
					)
				}
			},
		)
	}

	updateUploadProgress := func(
		progress r15m.UploadProgress,
	) {
		fyne.Do(
			func() {
				value :=
					float64(
						progress.Percent,
					) / 100.0

				if value < 0 {
					value = 0
				}

				if value > 1 {
					value = 1
				}

				progressBar.SetValue(value)

				progressLabel.SetText(
					fmt.Sprintf(
						"Enviando %d%%  (%s / %s)",
						progress.Percent,
						formatSize(
							progress.BytesSent,
						),
						formatSize(
							progress.TotalBytes,
						),
					),
				)
				progressLabel.Show()
				progressBar.Show()
			},
		)
	}

	selectGalleryImage :=
		func(item gallery.Item) {
			selected = item
			customSelected = false
			customInfo = nil

			preview.File =
				item.PreviewPath
			preview.Resource = nil
			preview.Image = nil
			preview.Refresh()

			selectedTitle.SetText(
				fmt.Sprintf(
					"Imagem %02d",
					item.ID,
				),
			)

			selectedInfo.SetText(
				fmt.Sprintf(
					"ACF: %s",
					formatSize(
						item.TextureSize,
					),
				),
			)

			sendButton.SetText(
				"Enviar para MiniTela",
			)
			sendButton.Enable()
			resetProgress()
		}

	sendButton.OnTapped = func() {
		if customSelected {
			info := customInfo

			if info == nil {
				dialog.ShowError(
					fmt.Errorf(
						"imagem própria não carregada",
					),
					w,
				)
				return
			}

			progressBar.SetValue(0)
			showProgress(
				"Gerando imagem...",
			)

			runBackground(
				"Gerando e enviando imagem própria...",
				"Imagem própria enviada.",
				func() error {
					fyne.Do(
						func() {
							setGalleryBusy(true)
						},
					)

					defer fyne.Do(
						func() {
							setGalleryBusy(false)
						},
					)

					templatePath,
						outputPath,
						err :=
						customimage.
							DefaultBuildPaths()

					if err != nil {
						return err
					}

					_, err =
						customimage.
							BuildStaticTextureFileWithProgress(
								info.Path,
								templatePath,
								outputPath,
								updateBuildStage,
							)

					if err != nil {
						return fmt.Errorf(
							"gerar ACF: %w",
							err,
						)
					}

					_, err =
						textureupload.
							SendFileWithStage(
								outputPath,
								updateUploadProgress,
								updateUploadStage,
							)

					if err != nil {
						fyne.Do(
							func() {
								progressLabel.SetText(
									"Falha no envio.",
								)
							},
						)
						return err
					}

					return nil
				},
			)

			return
		}

		item := selected

		progressBar.SetValue(0)
		showProgress(
			"Preparando upload...",
		)

		runBackground(
			fmt.Sprintf(
				"Enviando imagem %d...",
				item.ID,
			),
			fmt.Sprintf(
				"Imagem %d enviada.",
				item.ID,
			),
			func() error {
				fyne.Do(
					func() {
						setGalleryBusy(true)
					},
				)

				defer fyne.Do(
					func() {
						setGalleryBusy(false)
					},
				)

				_, err :=
					textureupload.
						SendFileWithStage(
							item.TexturePath,
							updateUploadProgress,
							updateUploadStage,
						)

				if err != nil {
					fyne.Do(
						func() {
							progressLabel.SetText(
								"Falha no envio.",
							)
						},
					)
				}

				return err
			},
		)
	}

	customButton =
		widget.NewButton(
			"+ Adicionar imagem própria",
			nil,
		)

	customButton.OnTapped = func() {
		fileDialog :=
			dialog.NewFileOpen(
				func(
					reader fyne.URIReadCloser,
					err error,
				) {
					if err != nil {
						dialog.ShowError(
							err,
							w,
						)
						return
					}

					if reader == nil {
						return
					}

					defer reader.Close()

					info, err :=
						customimage.Import(
							reader,
							reader.URI().Name(),
						)

					if err != nil {
						dialog.ShowError(
							err,
							w,
						)
						return
					}

					customSelected = true
					customInfo = info

					preview.File =
						info.Path
					preview.Resource = nil
					preview.Image = nil
					preview.Refresh()

					selectedTitle.SetText(
						"Imagem própria",
					)

					details :=
						fmt.Sprintf(
							"%s • %dx%d • %s",
							info.Format,
							info.Width,
							info.Height,
							formatSize(
								info.Size,
							),
						)

					if info.Frames > 1 {
						details +=
							fmt.Sprintf(
								" • %d frames",
								info.Frames,
							)
					}

					selectedInfo.SetText(
						details,
					)

					sendButton.SetText(
						"Gerar e enviar para MiniTela",
					)
					sendButton.Enable()
					resetProgress()
				},
				w,
			)

		fileDialog.SetFilter(
			storage.NewExtensionFileFilter(
				[]string{
					".png",
					".jpg",
					".jpeg",
					".gif",
				},
			),
		)

		fileDialog.Resize(
			fyne.NewSize(
				760,
				520,
			),
		)

		fileDialog.Show()
	}

	details :=
		widget.NewCard(
			"Selecionada",
			"Prévia da imagem",
			container.NewVBox(
				preview,
				selectedTitle,
				selectedInfo,
				progressLabel,
				progressBar,
				widget.NewSeparator(),
				sendButton,
				customButton,
			),
		)

	thumbs :=
		make(
			[]fyne.CanvasObject,
			0,
			len(items),
		)

	for _, item := range items {
		item := item

		thumb :=
			canvas.NewImageFromFile(
				item.PreviewPath,
			)
		thumb.FillMode =
			canvas.ImageFillContain
		thumb.CornerRadius = 8
		thumb.SetMinSize(
			fyne.NewSize(
				100,
				100,
			),
		)

		selectButton :=
			widget.NewButton(
				fmt.Sprintf(
					"Selecionar %02d",
					item.ID,
				),
				func() {
					selectGalleryImage(
						item,
					)
				},
			)
		selectButton.Importance =
			widget.LowImportance

		galleryButtons = append(
			galleryButtons,
			selectButton,
		)

		card :=
			widget.NewCard(
				fmt.Sprintf(
					"Imagem %02d",
					item.ID,
				),
				formatSize(
					item.TextureSize,
				),
				container.NewVBox(
					thumb,
					selectButton,
				),
			)

		thumbs = append(
			thumbs,
			card,
		)
	}

	grid :=
		container.NewGridWithColumns(
			4,
			thumbs...,
		)

	galleryScroll :=
		container.NewVScroll(
			grid,
		)

	split :=
		container.NewHSplit(
			galleryScroll,
			details,
		)
	split.SetOffset(0.68)

	header :=
		container.NewVBox(
			widget.NewLabel(
				fmt.Sprintf(
					"%d animações disponíveis",
					len(items),
				),
			),
			widget.NewSeparator(),
		)

	return container.NewBorder(
		header,
		nil,
		nil,
		nil,
		split,
	)
}

func buildCustomOnlyGallery(
	w fyne.Window,
	runBackground func(
		string,
		string,
		func() error,
	),
) fyne.CanvasObject {
	var customInfo *customimage.Info

	preview :=
		canvas.NewImageFromFile("")
	preview.FillMode =
		canvas.ImageFillContain
	preview.CornerRadius = 12
	preview.SetMinSize(
		fyne.NewSize(
			280,
			280,
		),
	)

	selectedTitle :=
		widget.NewLabelWithStyle(
			"Nenhuma imagem selecionada",
			fyne.TextAlignCenter,
			fyne.TextStyle{
				Bold: true,
			},
		)

	selectedInfo :=
		widget.NewLabel(
			"Selecione uma imagem PNG, JPG ou GIF.",
		)
	selectedInfo.Alignment =
		fyne.TextAlignCenter
	selectedInfo.Wrapping =
		fyne.TextWrapWord

	progressLabel :=
		widget.NewLabel("")
	progressLabel.Alignment =
		fyne.TextAlignCenter
	progressLabel.Hide()

	progressBar :=
		widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 1
	progressBar.SetValue(0)
	progressBar.Hide()

	showProgress := func(text string) {
		progressLabel.SetText(text)
		progressLabel.Show()
		progressBar.Show()
	}

	updateBuildStage := func(
		stage customimage.BuildStage,
	) {
		fyne.Do(
			func() {
				switch stage {
				case customimage.BuildStageImage:
					showProgress(
						"Gerando imagem...",
					)

				case customimage.BuildStageSTCRGBA:
					showProgress(
						"Convertendo STCRGBA...",
					)

				case customimage.BuildStageACF:
					showProgress(
						"Preparando ACF...",
					)
				}
			},
		)
	}

	updateUploadStage := func(
		stage textureupload.Stage,
	) {
		fyne.Do(
			func() {
				switch stage {
				case textureupload.StagePreparing:
					showProgress(
						"Preparando upload...",
					)

				case textureupload.StageConnecting:
					showProgress(
						"Conectando à MiniTela...",
					)

				case textureupload.StageUploading:
					progressBar.SetValue(0)
					showProgress(
						"Enviando 0%",
					)

				case textureupload.StageReconnecting:
					progressBar.SetValue(1)
					showProgress(
						"Reconectando MiniTela...",
					)

				case textureupload.StageSelectingScreen:
					progressBar.SetValue(1)
					showProgress(
						"Ativando imagem na MiniTela...",
					)

				case textureupload.StageDone:
					progressBar.SetValue(1)
					showProgress(
						"Imagem enviada ✓",
					)
				}
			},
		)
	}

	updateUploadProgress := func(
		progress r15m.UploadProgress,
	) {
		fyne.Do(
			func() {
				value :=
					float64(
						progress.Percent,
					) / 100.0

				if value < 0 {
					value = 0
				}

				if value > 1 {
					value = 1
				}

				progressBar.SetValue(value)
				progressLabel.SetText(
					fmt.Sprintf(
						"Enviando %d%%  (%s / %s)",
						progress.Percent,
						formatSize(
							progress.BytesSent,
						),
						formatSize(
							progress.TotalBytes,
						),
					),
				)
				progressLabel.Show()
				progressBar.Show()
			},
		)
	}

	sendButton :=
		widget.NewButton(
			"Gerar e enviar para MiniTela",
			nil,
		)
	sendButton.Importance =
		widget.HighImportance
	sendButton.Disable()

	var imageButton *widget.Button

	setBusy := func(value bool) {
		if value {
			sendButton.Disable()

			if imageButton != nil {
				imageButton.Disable()
			}

			return
		}

		if customInfo != nil {
			sendButton.Enable()
		}

		if imageButton != nil {
			imageButton.Enable()
		}

	}

	sendButton.OnTapped = func() {
		info := customInfo

		if info == nil {
			dialog.ShowInformation(
				"Minha imagem",
				"Selecione uma imagem antes de enviar.",
				w,
			)
			return
		}

		if _, err :=
			customimage.ResolveTemplatePath(); err != nil {
			dialog.ShowError(
				fmt.Errorf(
					"componente interno de imagem indisponível: %w",
					err,
				),
				w,
			)
			return
		}

		progressBar.SetValue(0)
		showProgress(
			"Gerando imagem...",
		)

		runBackground(
			"Gerando e enviando imagem própria...",
			"Imagem própria enviada.",
			func() error {
				fyne.Do(
					func() {
						setBusy(true)
					},
				)

				defer fyne.Do(
					func() {
						setBusy(false)
					},
				)

				templatePath,
					outputPath,
					err :=
					customimage.
						DefaultBuildPaths()

				if err != nil {
					return err
				}

				_, err =
					customimage.
						BuildStaticTextureFileWithProgress(
							info.Path,
							templatePath,
							outputPath,
							updateBuildStage,
						)

				if err != nil {
					return fmt.Errorf(
						"gerar ACF: %w",
						err,
					)
				}

				_, err =
					textureupload.
						SendFileWithStage(
							outputPath,
							updateUploadProgress,
							updateUploadStage,
						)

				if err != nil {
					fyne.Do(
						func() {
							progressLabel.SetText(
								"Falha no envio.",
							)
						},
					)
				}

				return err
			},
		)
	}

	imageButton =
		widget.NewButton(
			"+ Selecionar imagem",
			func() {
				fileDialog :=
					dialog.NewFileOpen(
						func(
							reader fyne.URIReadCloser,
							err error,
						) {
							if err != nil {
								dialog.ShowError(
									err,
									w,
								)
								return
							}

							if reader == nil {
								return
							}

							defer reader.Close()

							info, err :=
								customimage.Import(
									reader,
									reader.URI().Name(),
								)

							if err != nil {
								dialog.ShowError(
									err,
									w,
								)
								return
							}

							customInfo = info

							preview.File =
								info.Path
							preview.Resource = nil
							preview.Image = nil
							preview.Refresh()

							selectedTitle.SetText(
								"Imagem própria",
							)

							details :=
								fmt.Sprintf(
									"%s • %dx%d • %s",
									info.Format,
									info.Width,
									info.Height,
									formatSize(
										info.Size,
									),
								)

							if info.Frames > 1 {
								details +=
									fmt.Sprintf(
										" • %d frames",
										info.Frames,
									)
							}

							selectedInfo.SetText(
								details,
							)
							progressBar.SetValue(0)
							progressLabel.Hide()
							progressBar.Hide()
							sendButton.Enable()
						},
						w,
					)

				fileDialog.SetFilter(
					storage.NewExtensionFileFilter(
						[]string{
							".png",
							".jpg",
							".jpeg",
							".gif",
						},
					),
				)

				fileDialog.Resize(
					fyne.NewSize(
						760,
						520,
					),
				)

				fileDialog.Show()
			},
		)

	customCard :=
		widget.NewCard(
			"Minha imagem",
			"PNG, JPG e GIF animado",
			container.NewVBox(
				preview,
				selectedTitle,
				selectedInfo,
				progressLabel,
				progressBar,
				widget.NewSeparator(),
				imageButton,
				sendButton,
			),
		)

	originalInfo :=
		widget.NewLabel(
			"A galeria original é opcional e não está instalada neste computador. " +
				"Você pode usar normalmente suas próprias imagens.",
		)
	originalInfo.Wrapping =
		fyne.TextWrapWord

	originalCard :=
		widget.NewCard(
			"Galeria original",
			"Recursos opcionais",
			originalInfo,
		)

	return container.NewGridWithColumns(
		2,
		customCard,
		originalCard,
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

type environmentCheck struct {
	DeviceDetected bool
	DevicePath     string
	DeviceError    string
	SystemdUser    bool
	ServiceActive  bool
	ServiceEnabled bool
	UdevRule       bool
	StableAlias    bool
	Pkexec         bool
}

func inspectEnvironment() environmentCheck {
	check := environmentCheck{
		SystemdUser: systemdUserAvailable(),
		UdevRule:    fileExists("/etc/udev/rules.d/99-minitela.rules"),
		StableAlias: fileExists("/dev/minitela-r15m"),
	}

	if _, err := exec.LookPath("pkexec"); err == nil {
		check.Pkexec = true
	}

	detected, err := device.DetectR15M()
	if err == nil {
		check.DeviceDetected = true
		check.DevicePath = detected.Path
	} else {
		check.DeviceError = err.Error()
	}

	if check.SystemdUser {
		check.ServiceActive = serviceActive()
		check.ServiceEnabled = serviceEnabled()
	}

	return check
}

func (check environmentCheck) primaryStatus() string {
	switch {
	case !check.SystemdUser:
		return "⚠ systemd do usuário indisponível"

	case !check.DeviceDetected:
		return "○ MiniTela não detectada"

	case check.ServiceActive:
		return "● MiniTela conectada / serviço ativo"

	default:
		return "○ MiniTela detectada / serviço parado"
	}
}

func (check environmentCheck) detailStatus() string {
	var parts []string

	if check.DeviceDetected {
		parts = append(
			parts,
			"Dispositivo: "+check.DevicePath,
		)
	} else {
		parts = append(
			parts,
			"Conecte o Positivo Vision R15M via USB.",
		)
	}

	if !check.SystemdUser {
		parts = append(
			parts,
			"Serviço de usuário do systemd não disponível.",
		)
	}

	if !check.UdevRule {
		if check.Pkexec {
			parts = append(
				parts,
				"Permissão udev pendente; reabra o AppImage e autorize a configuração.",
			)
		} else {
			parts = append(
				parts,
				"Permissão udev pendente e pkexec não está disponível.",
			)
		}
	} else if !check.StableAlias && check.DeviceDetected {
		parts = append(
			parts,
			"Regra udev instalada; alias /dev/minitela-r15m ainda não apareceu.",
		)
	}

	return strings.Join(parts, " • ")
}

func (check environmentCheck) report() string {
	deviceLine := "não detectada"
	if check.DeviceDetected {
		deviceLine = "detectada em " + check.DevicePath
	} else if check.DeviceError != "" {
		deviceLine += " (" + check.DeviceError + ")"
	}

	lines := []string{
		"R15M: " + deviceLine,
		"systemd --user: " + availableText(check.SystemdUser),
		"Backend: " + activeText(check.ServiceActive),
		"Autostart: " + enabledText(check.ServiceEnabled),
		"Regra udev: " + installedText(check.UdevRule),
		"Alias /dev/minitela-r15m: " + availableText(check.StableAlias),
		"pkexec: " + availableText(check.Pkexec),
	}

	if setup := lastSetupLog(); setup != "" {
		lines = append(
			lines,
			"",
			"Último setup:",
			setup,
		)
	}

	return strings.Join(lines, "\n")
}

func systemdUserAvailable() bool {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return false
	}

	command :=
		exec.Command(
			"systemctl",
			"--user",
			"show-environment",
		)

	return command.Run() == nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func lastSetupLog() string {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}

		stateDir =
			filepath.Join(
				home,
				".local",
				"state",
			)
	}

	data, err :=
		os.ReadFile(
			filepath.Join(
				stateDir,
				"minitela",
				"setup.log",
			),
		)

	if err != nil {
		return ""
	}

	text :=
		strings.TrimSpace(
			string(data),
		)

	if text == "" {
		return ""
	}

	lines :=
		strings.Split(
			text,
			"\n",
		)

	if len(lines) > 5 {
		lines = lines[len(lines)-5:]
	}

	return strings.Join(lines, "\n")
}

func availableText(value bool) string {
	if value {
		return "disponível"
	}

	return "indisponível"
}

func activeText(value bool) string {
	if value {
		return "ativo"
	}

	return "parado"
}

func enabledText(value bool) string {
	if value {
		return "ativado"
	}

	return "desativado"
}

func installedText(value bool) string {
	if value {
		return "instalada"
	}

	return "ausente"
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
