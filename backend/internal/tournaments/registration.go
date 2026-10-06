package tournaments

func normalizeFormat(format string) string {
	if format == "" {
		return "swiss"
	}
	return format
}
