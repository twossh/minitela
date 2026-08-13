package r15m

import "fmt"

func BatteryLevel(percent int) uint32 {
	if percent <= 0 {
		return 0
	}

	level := percent / 25

	if level > 3 {
		level = 3
	}

	return uint32(level)
}

func BatteryText(percent int) []byte {
	if percent < 0 {
		percent = 0
	}

	if percent > 100 {
		percent = 100
	}

	return []byte(
		fmt.Sprintf("%d%%", percent),
	)
}
