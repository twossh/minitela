package r15m

import "testing"

func TestDisplayText(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"WiFi",
			"WiFi",
		},
		{
			"Honda Nicola 5G",
			"Honda Nic...",
		},
	}

	for _, tt := range tests {
		got := DisplayText(tt.input)

		if got != tt.want {
			t.Fatalf(
				"DisplayText(%q)=%q esperado=%q",
				tt.input,
				got,
				tt.want,
			)
		}
	}
}
