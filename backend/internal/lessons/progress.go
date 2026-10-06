package lessons

func courseProgress(chapters []Chapter) int {
	if len(chapters) == 0 {
		return 0
	}
	total := 0
	for _, chapter := range chapters {
		total += chapter.ProgressPercent
	}
	return total / len(chapters)
}
