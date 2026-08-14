package r15m

import (
	"fmt"
	"strings"
)

type Screen uint32

const (
	ScreenWhatsApp Screen = 1
	ScreenNotes    Screen = 2
	ScreenMonitor  Screen = 3
	ScreenWeather  Screen = 4
	ScreenImage    Screen = 5
)

func (s Screen) String() string {
	switch s {
	case ScreenWhatsApp:
		return "WhatsApp"
	case ScreenNotes:
		return "Notas"
	case ScreenMonitor:
		return "Monitor"
	case ScreenWeather:
		return "Clima"
	case ScreenImage:
		return "Imagem"
	default:
		return fmt.Sprintf("Página %d", s)
	}
}

func ParseScreen(value string) (Screen, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "whatsapp":
		return ScreenWhatsApp, nil

	case "notes", "notas":
		return ScreenNotes, nil

	case "monitor":
		return ScreenMonitor, nil

	case "weather", "clima":
		return ScreenWeather, nil

	case "image", "imagem":
		return ScreenImage, nil

	default:
		return 0, fmt.Errorf(
			"tela desconhecida %q; use whatsapp, notes, monitor, weather ou image",
			value,
		)
	}
}

func (c *Connection) SetScreen(
	screen Screen,
) error {
	if screen < ScreenWhatsApp ||
		screen > ScreenImage {
		return fmt.Errorf(
			"tela inválida: %d",
			screen,
		)
	}

	actual, err := c.WriteRegisterVerified(
		RegisterCurrentPage,
		uint32(screen),
	)
	if err != nil {
		return err
	}

	if actual != uint32(screen) {
		return fmt.Errorf(
			"tela solicitada=%d confirmada=%d",
			screen,
			actual,
		)
	}

	return nil
}
