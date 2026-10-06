package chat

import (
	"context"
	"strings"
	"unicode"
)

type BasicModeration struct{}

func (BasicModeration) Check(_ context.Context, _ string, _ string, content string) error {
	if len(strings.TrimSpace(content)) == 0 || len(content) > 2000 {
		return ErrInvalid
	}
	for _, char := range content {
		if unicode.IsControl(char) && char != '\n' && char != '\t' {
			return ErrModerated
		}
	}
	return nil
}
