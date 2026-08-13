package r15m

func DisplayText(value string) string {
	runes := []rune(value)

	if len(runes) <= 12 {
		return value
	}

	return string(runes[:9]) + "..."
}
