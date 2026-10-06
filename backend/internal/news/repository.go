package news

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

const postColumns = `
	SELECT id::text, author_id::text, title, slug, excerpt, body, COALESCE(cover_url, ''),
		tags, status, published_at, created_at, updated_at FROM news_posts`

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]Post, error) {
	rows, err := r.db.QueryContext(ctx, postColumns+`
		WHERE status = 'published' AND published_at <= now()
		ORDER BY published_at DESC, id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list published news: %w", err)
	}
	defer rows.Close()
	items := make([]Post, 0)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate news posts: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ListAdmin(ctx context.Context, status string, limit, offset int) ([]Post, error) {
	rows, err := r.db.QueryContext(ctx, postColumns+`
		WHERE $1 = '' OR status = $1
		ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list admin news posts: %w", err)
	}
	defer rows.Close()
	items := make([]Post, 0)
	for rows.Next() {
		item, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin news posts: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) Get(ctx context.Context, slug string) (Post, error) {
	item, err := scanPost(r.db.QueryRowContext(ctx, postColumns+`
		WHERE slug = $1 AND status = 'published' AND published_at <= now()`, slug))
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) Create(ctx context.Context, authorID string, input Input) (Post, error) {
	tags, err := json.Marshal(input.Tags)
	if err != nil {
		return Post{}, fmt.Errorf("encode news tags: %w", err)
	}
	item, err := scanPost(r.db.QueryRowContext(ctx, `
		INSERT INTO news_posts (author_id, title, slug, excerpt, body, cover_url, tags)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7)
		RETURNING id::text, author_id::text, title, slug, excerpt, body, COALESCE(cover_url, ''),
			tags, status, published_at, created_at, updated_at`,
		authorID, input.Title, input.Slug, input.Excerpt, input.Body, input.CoverURL, tags))
	if isSlugConflict(err) {
		return Post{}, ErrConflict
	}
	return item, err
}

func (r *PostgresRepository) Update(ctx context.Context, postID string, input Input) (Post, error) {
	tags, err := json.Marshal(input.Tags)
	if err != nil {
		return Post{}, fmt.Errorf("encode news tags: %w", err)
	}
	item, err := scanPost(r.db.QueryRowContext(ctx, `
		UPDATE news_posts SET title = $2, slug = $3, excerpt = $4, body = $5,
			cover_url = NULLIF($6, ''), tags = $7, updated_at = now()
		WHERE id = $1
		RETURNING id::text, author_id::text, title, slug, excerpt, body, COALESCE(cover_url, ''),
			tags, status, published_at, created_at, updated_at`,
		postID, input.Title, input.Slug, input.Excerpt, input.Body, input.CoverURL, tags))
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	if isSlugConflict(err) {
		return Post{}, ErrConflict
	}
	return item, err
}

func isSlugConflict(err error) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == "23505" && pgError.ConstraintName == "news_posts_slug_key"
}

func (r *PostgresRepository) SetStatus(ctx context.Context, postID, status string) (Post, error) {
	item, err := scanPost(r.db.QueryRowContext(ctx, `
		UPDATE news_posts SET status = $2,
			published_at = CASE WHEN $2 = 'published' THEN COALESCE(published_at, now()) ELSE NULL END,
			updated_at = now()
		WHERE id = $1
		RETURNING id::text, author_id::text, title, slug, excerpt, body, COALESCE(cover_url, ''),
			tags, status, published_at, created_at, updated_at`, postID, status))
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) Delete(ctx context.Context, postID string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM news_posts WHERE id = $1`, postID)
	if err != nil {
		return fmt.Errorf("delete news post: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read news deletion result: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanPost(row scanner) (Post, error) {
	var item Post
	var tags []byte
	err := row.Scan(&item.ID, &item.AuthorID, &item.Title, &item.Slug, &item.Excerpt, &item.Body,
		&item.CoverURL, &tags, &item.Status, &item.PublishedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Post{}, fmt.Errorf("read news post: %w", err)
	}
	if err := json.Unmarshal(tags, &item.Tags); err != nil {
		return Post{}, fmt.Errorf("decode news tags: %w", err)
	}
	return item, nil
}

var _ Repository = (*PostgresRepository)(nil)
