package r15m

import "testing"

func TestBatteryLevel(t *testing.T) {
	tests := []struct {
		percent int
		want    uint32
	}{
		{0, 0},
		{1, 0},
		{24, 0},

		{25, 1},
		{49, 1},

		{50, 2},
		{74, 2},

		{75, 3},
		{99, 3},
		{100, 3},
	}

	for _, tt := range tests {
		got := BatteryLevel(tt.percent)

		if got != tt.want {
			t.Fatalf(
				"BatteryLevel(%d)=%d esperado=%d",
				tt.percent,
				got,
				tt.want,
			)
		}
	}
}

func TestBatteryText(t *testing.T) {
	got := string(BatteryText(82))

	if got != "82%" {
		t.Fatalf(
			"BatteryText=%q esperado=\"82%%\"",
			got,
		)
	}
}
