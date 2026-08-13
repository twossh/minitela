package r15m

import "testing"

func TestParseScreen(t *testing.T) {
	tests := []struct {
		input string
		want  Screen
	}{
		{"whatsapp", ScreenWhatsApp},
		{"notes", ScreenNotes},
		{"notas", ScreenNotes},
		{"monitor", ScreenMonitor},
		{"weather", ScreenWeather},
		{"clima", ScreenWeather},
	}

	for _, tt := range tests {
		got, err := ParseScreen(tt.input)
		if err != nil {
			t.Fatalf(
				"ParseScreen(%q): %v",
				tt.input,
				err,
			)
		}

		if got != tt.want {
			t.Fatalf(
				"ParseScreen(%q)=%d, esperado=%d",
				tt.input,
				got,
				tt.want,
			)
		}
	}
}

func TestScreenString(t *testing.T) {
	tests := []struct {
		screen Screen
		want   string
	}{
		{ScreenWhatsApp, "WhatsApp"},
		{ScreenNotes, "Notas"},
		{ScreenMonitor, "Monitor"},
		{ScreenWeather, "Clima"},
	}

	for _, tt := range tests {
		if got := tt.screen.String(); got != tt.want {
			t.Fatalf(
				"Screen(%d).String()=%q, esperado=%q",
				tt.screen,
				got,
				tt.want,
			)
		}
	}
}
