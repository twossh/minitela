package weather

import (
	"testing"
	"time"
)

func TestR15MIcon(
	t *testing.T,
) {
	tests := []struct {
		condition string
		want      uint32
	}{
		{"Sunny", 0},
		{"Clear", 0},
		{"Partly cloudy", 1},
		{"Cloudy", 2},
		{"Overcast", 2},
		{"Mist", 2},
		{"Patchy rain nearby", 3},
		{"Light drizzle", 3},
		{"Thundery outbreaks", 3},
		{"Light snow", 6},
		{"Blizzard", 6},
	}

	for _, tt := range tests {
		got :=
			R15MIcon(
				tt.condition,
			)

		if got != tt.want {
			t.Fatalf(
				"R15MIcon(%q)=%d esperado=%d",
				tt.condition,
				got,
				tt.want,
			)
		}
	}
}

func TestTemperatureText(
	t *testing.T,
) {
	got :=
		TemperatureText(
			17,
			25,
		)

	if got != "17°/25°" {
		t.Fatalf(
			"TemperatureText=%q",
			got,
		)
	}
}

func TestDayText(
	t *testing.T,
) {
	date :=
		time.Date(
			2026,
			time.August,
			14,
			0,
			0,
			0,
			0,
			time.UTC,
		)

	got :=
		DayText(date)

	if got != "SEX14" {
		t.Fatalf(
			"DayText=%q",
			got,
		)
	}
}

func TestCityText(
	t *testing.T,
) {
	got :=
		CityText(
			"Porto Alegre",
		)

	if got != "Porto Al..." {
		t.Fatalf(
			"CityText=%q",
			got,
		)
	}
}
