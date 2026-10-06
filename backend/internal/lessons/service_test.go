package lessons

import (
	"context"
	"encoding/json"
	"testing"
)

const lessonUserID = "11111111-1111-4111-8111-111111111111"
const lessonChapterID = "22222222-2222-4222-8222-222222222222"

type fakeRepository struct {
	input ProgressInput
}

func (r *fakeRepository) List(context.Context, string) ([]Course, error) {
	return nil, nil
}

func (r *fakeRepository) Get(context.Context, string, string) (Course, error) {
	return Course{}, nil
}

func (r *fakeRepository) UpdateProgress(_ context.Context, _, _ string, input ProgressInput) (Chapter, error) {
	r.input = input
	return Chapter{ProgressPercent: input.ProgressPercent, Completed: input.Completed}, nil
}

func TestUpdateProgressCompletionAndValidation(t *testing.T) {
	repository := new(fakeRepository)
	service := NewService(repository)
	chapter, err := service.UpdateProgress(context.Background(), lessonUserID, lessonChapterID, ProgressInput{
		ProgressPercent: 70,
		Completed:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if chapter.ProgressPercent != 100 || !repository.input.Completed {
		t.Fatalf("completion not normalized: %+v", repository.input)
	}

	_, err = service.UpdateProgress(context.Background(), lessonUserID, lessonChapterID, ProgressInput{
		ProgressPercent: 101,
	})
	if err == nil {
		t.Fatal("out-of-range progress accepted")
	}
	_, err = service.UpdateProgress(context.Background(), lessonUserID, lessonChapterID, ProgressInput{
		LastPosition: json.RawMessage(`[]`),
	})
	if err == nil {
		t.Fatal("non-object last position accepted")
	}
}

func TestCourseProgress(t *testing.T) {
	got := courseProgress([]Chapter{{ProgressPercent: 100}, {ProgressPercent: 50}, {ProgressPercent: 0}})
	if got != 50 {
		t.Fatalf("course progress = %d, want 50", got)
	}
	if got := courseProgress(nil); got != 0 {
		t.Fatalf("empty course progress = %d, want 0", got)
	}
}
