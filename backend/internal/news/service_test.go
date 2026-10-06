package news

import (
	"testing"
)

func TestNormalizeGeneratesSlugAndTrimsTags(t *testing.T) {
	input, err := normalize(Input{
		Title: "  Chess News: A New Era! ",
		Body:  "  Article content ",
		Tags:  []string{" tactics ", "analysis"},
	})
	if err != nil {
		t.Fatalf("normalize() error = %v", err)
	}
	if input.Slug != "chess-news-a-new-era" || input.Body != "Article content" || input.Tags[0] != "tactics" {
		t.Fatalf("normalized input = %+v", input)
	}
}

func TestNormalizeRejectsInvalidSlug(t *testing.T) {
	_, err := normalize(Input{Title: "Post", Slug: "bad/path", Body: "Body"})
	if err != ErrInvalid {
		t.Fatalf("normalize() error = %v, want ErrInvalid", err)
	}
}
