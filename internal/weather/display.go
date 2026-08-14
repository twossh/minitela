package weather

import (
	"fmt"
	"strings"
	"time"
)

func R15MIcon(condition string) uint32 {
	value := strings.ToLower(
		strings.TrimSpace(condition),
	)

	switch {
	case strings.Contains(value, "sunny"),
		strings.Contains(value, "clear"):
		return 0

	case strings.Contains(value, "partly cloudy"):
		return 1

	case strings.Contains(value, "snow"),
		strings.Contains(value, "ice"),
		strings.Contains(value, "blizzard"),
		strings.Contains(value, "freezing"),
		strings.Contains(value, "sleet"):
		return 6

	case strings.Contains(value, "rain"),
		strings.Contains(value, "drizzle"),
		strings.Contains(value, "thunder"),
		strings.Contains(value, "shower"):
		return 3

	case strings.Contains(value, "cloudy"),
		strings.Contains(value, "overcast"),
		strings.Contains(value, "fog"),
		strings.Contains(value, "mist"):
		return 2

	default:
		return 2
	}
}

func TemperatureText(
	minTemp int,
	maxTemp int,
) string {
	return fmt.Sprintf(
		"%d°/%d°",
		minTemp,
		maxTemp,
	)
}

func DayText(date time.Time) string {
	weekday := map[time.Weekday]string{
		time.Sunday:    "DOM",
		time.Monday:    "SEG",
		time.Tuesday:   "TER",
		time.Wednesday: "QUA",
		time.Thursday:  "QUI",
		time.Friday:    "SEX",
		time.Saturday:  "SÁB",
	}

	return fmt.Sprintf(
		"%s%02d",
		weekday[date.Weekday()],
		date.Day(),
	)
}

func CityText(value string) string {
	value = strings.TrimSpace(value)

	runes := []rune(value)

	if len(runes) <= 10 {
		return value
	}

	return string(runes[:8]) + "..."
}
