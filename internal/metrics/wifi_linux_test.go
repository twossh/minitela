//go:build linux

package metrics

import "testing"

func TestSignalQuality(t *testing.T) {
	tests := []struct {
		dbm  int
		want int
	}{
		{-110, 0},
		{-100, 0},
		{-90, 20},
		{-75, 50},
		{-60, 80},
		{-50, 100},
		{-40, 100},
	}

	for _, tt := range tests {
		got := SignalQuality(tt.dbm)

		if got != tt.want {
			t.Fatalf(
				"SignalQuality(%d)=%d esperado=%d",
				tt.dbm,
				got,
				tt.want,
			)
		}
	}
}
