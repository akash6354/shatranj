package lessons

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context, userID string) ([]Course, error) {
	rows, err := r.db.QueryContext(ctx, lessonSelect+`
		WHERE c.status = 'published'
		ORDER BY c.sort_order, c.title, ch.sort_order, ch.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list lessons: %w", err)
	}
	defer rows.Close()
	courses := make([]Course, 0)
	positions := make(map[string]int)
	for rows.Next() {
		course, chapter, hasChapter, err := scanLessonRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan lesson: %w", err)
		}
		position, found := positions[course.ID]
		if !found {
			course.Chapters = make([]Chapter, 0)
			positions[course.ID] = len(courses)
			courses = append(courses, course)
			position = len(courses) - 1
		}
		if hasChapter {
			courses[position].Chapters = append(courses[position].Chapters, chapter)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate lessons: %w", err)
	}
	for index := range courses {
		courses[index].ProgressPercent = courseProgress(courses[index].Chapters)
	}
	return courses, nil
}

func (r *PostgresRepository) Get(ctx context.Context, userID, courseID string) (Course, error) {
	rows, err := r.db.QueryContext(ctx, lessonSelect+`
		WHERE c.id = $2 AND c.status = 'published'
		ORDER BY ch.sort_order, ch.id`, userID, courseID)
	if err != nil {
		return Course{}, fmt.Errorf("get lesson: %w", err)
	}
	defer rows.Close()
	var course Course
	found := false
	course.Chapters = make([]Chapter, 0)
	for rows.Next() {
		rowCourse, chapter, hasChapter, err := scanLessonRow(rows)
		if err != nil {
			return Course{}, fmt.Errorf("scan lesson: %w", err)
		}
		if !found {
			course = rowCourse
			course.Chapters = make([]Chapter, 0)
			found = true
		}
		if hasChapter {
			course.Chapters = append(course.Chapters, chapter)
		}
	}
	if err := rows.Err(); err != nil {
		return Course{}, fmt.Errorf("iterate lesson chapters: %w", err)
	}
	if !found {
		return Course{}, ErrNotFound
	}
	course.ProgressPercent = courseProgress(course.Chapters)
	return course, nil
}

func (r *PostgresRepository) UpdateProgress(ctx context.Context, userID, chapterID string, input ProgressInput) (Chapter, error) {
	lastPosition := input.LastPosition
	if len(lastPosition) == 0 {
		lastPosition = json.RawMessage(`{}`)
	}
	var chapter Chapter
	var content []byte
	var storedPosition []byte
	var updatedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO user_lesson_progress
			(user_id, chapter_id, progress_percent, completed, last_position, updated_at)
		SELECT $1, ch.id, $3, $4, $5::jsonb, now()
		FROM lesson_chapters ch
		JOIN lesson_courses c ON c.id = ch.course_id
		WHERE ch.id = $2 AND c.status = 'published'
		ON CONFLICT (user_id, chapter_id) DO UPDATE
		SET progress_percent = GREATEST(user_lesson_progress.progress_percent, EXCLUDED.progress_percent),
			completed = user_lesson_progress.completed OR EXCLUDED.completed,
			last_position = CASE WHEN $6 THEN EXCLUDED.last_position ELSE user_lesson_progress.last_position END,
			updated_at = now()
		RETURNING chapter_id, progress_percent, completed, last_position, updated_at`,
		userID, chapterID, input.ProgressPercent, input.Completed, string(lastPosition), len(input.LastPosition) > 0).
		Scan(&chapter.ID, &chapter.ProgressPercent, &chapter.Completed, &storedPosition, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Chapter{}, ErrNotFound
	}
	if err != nil {
		return Chapter{}, fmt.Errorf("update lesson progress: %w", err)
	}
	if updatedAt.Valid {
		chapter.UpdatedAt = &updatedAt.Time
	}
	chapter.LastPosition = json.RawMessage(storedPosition)
	if err := r.db.QueryRowContext(ctx, `
		SELECT course_id, title, description, sort_order, content
		FROM lesson_chapters WHERE id = $1`, chapter.ID).Scan(
		&chapter.CourseID, &chapter.Title, &chapter.Description, &chapter.SortOrder, &content); err != nil {
		return Chapter{}, fmt.Errorf("load updated lesson chapter: %w", err)
	}
	chapter.Content = json.RawMessage(content)
	return chapter, nil
}

const lessonSelect = `
	SELECT c.id, c.title, c.description, c.category, c.cover_image_url, c.sort_order,
		c.created_at, c.updated_at,
		ch.id, ch.title, ch.description, ch.sort_order, ch.content,
		COALESCE(p.progress_percent, 0), COALESCE(p.completed, FALSE),
		COALESCE(p.last_position, '{}'::jsonb), p.updated_at
	FROM lesson_courses c
	LEFT JOIN lesson_chapters ch ON ch.course_id = c.id
	LEFT JOIN user_lesson_progress p ON p.chapter_id = ch.id AND p.user_id = $1
`

type rowScanner interface {
	Scan(...any) error
}

func scanLessonRow(row rowScanner) (Course, Chapter, bool, error) {
	var course Course
	var chapter Chapter
	var chapterID, chapterTitle, chapterDescription sql.NullString
	var chapterOrder sql.NullInt64
	var content []byte
	var progress sql.NullInt64
	var completed sql.NullBool
	var lastPosition []byte
	var progressUpdatedAt sql.NullTime
	err := row.Scan(&course.ID, &course.Title, &course.Description, &course.Category,
		&course.CoverImageURL, &course.SortOrder, &course.CreatedAt, &course.UpdatedAt,
		&chapterID, &chapterTitle, &chapterDescription, &chapterOrder, &content,
		&progress, &completed, &lastPosition, &progressUpdatedAt)
	if err != nil {
		return Course{}, Chapter{}, false, err
	}
	if !chapterID.Valid {
		return course, Chapter{}, false, nil
	}
	chapter.ID = chapterID.String
	chapter.CourseID = course.ID
	chapter.Title = chapterTitle.String
	chapter.Description = chapterDescription.String
	chapter.SortOrder = int(chapterOrder.Int64)
	chapter.Content = json.RawMessage(content)
	chapter.ProgressPercent = int(progress.Int64)
	chapter.Completed = completed.Bool
	chapter.LastPosition = json.RawMessage(lastPosition)
	if progressUpdatedAt.Valid {
		chapter.UpdatedAt = &progressUpdatedAt.Time
	}
	return course, chapter, true, nil
}
