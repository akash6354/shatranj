package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) Get(ctx context.Context, userID string) (Subscription, error) {
	var subscription Subscription
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id::text, plan_id, status, started_at, ends_at, updated_at
		FROM user_subscriptions WHERE user_id = $1`, userID).Scan(
		&subscription.UserID, &subscription.PlanID, &subscription.Status,
		&subscription.StartedAt, &subscription.EndsAt, &subscription.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Subscription{UserID: userID, PlanID: PlanFree, Status: "active"}, nil
	}
	if err != nil {
		return Subscription{}, fmt.Errorf("get subscription: %w", err)
	}
	if subscription.PlanID == PlanPremium && subscription.EndsAt != nil && !subscription.EndsAt.After(time.Now()) {
		subscription.Status = "expired"
	}
	return subscription, nil
}

func (r *PostgresRepository) ActivatePremium(ctx context.Context, userID string, start, end time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_subscriptions (user_id, plan_id, status, started_at, ends_at, updated_at)
		VALUES ($1, 'premium', 'active', $2, $3, now())
		ON CONFLICT (user_id) DO UPDATE SET plan_id = 'premium', status = 'active',
			started_at = EXCLUDED.started_at, ends_at = EXCLUDED.ends_at, updated_at = now()`,
		userID, start, end)
	if err != nil {
		return fmt.Errorf("activate premium subscription: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ActivateForPayment(ctx context.Context, userID string, start, end time.Time, paymentID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin payment subscription activation: %w", err)
	}
	defer tx.Rollback()
	var validPayment bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM payments WHERE id = $1 AND user_id = $2 AND status = 'paid')`,
		paymentID, userID).Scan(&validPayment); err != nil {
		return fmt.Errorf("verify paid subscription order: %w", err)
	}
	if !validPayment {
		return ErrInvalid
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO subscription_payment_activations (payment_id, user_id)
		VALUES ($1, $2) ON CONFLICT (payment_id) DO NOTHING`, paymentID, userID)
	if err != nil {
		return fmt.Errorf("record subscription payment activation: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read subscription payment activation result: %w", err)
	}
	if inserted == 0 {
		return tx.Commit()
	}

	var currentEnd sql.NullTime
	if err := tx.QueryRowContext(ctx, `
		SELECT ends_at FROM user_subscriptions
		WHERE user_id = $1 AND plan_id = 'premium' AND status = 'active' AND ends_at > $2
		FOR UPDATE`, userID, start).Scan(&currentEnd); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read current premium term: %w", err)
	}
	if currentEnd.Valid {
		end = currentEnd.Time.Add(end.Sub(start))
		start = currentEnd.Time
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO user_subscriptions (user_id, plan_id, status, started_at, ends_at, updated_at)
		VALUES ($1, 'premium', 'active', $2, $3, now())
		ON CONFLICT (user_id) DO UPDATE SET plan_id = 'premium', status = 'active',
			started_at = EXCLUDED.started_at, ends_at = EXCLUDED.ends_at, updated_at = now()`,
		userID, start, end); err != nil {
		return fmt.Errorf("apply premium term: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit payment subscription activation: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Cancel(ctx context.Context, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE user_subscriptions SET status = 'cancelled', updated_at = now()
		WHERE user_id = $1 AND plan_id = 'premium' AND status = 'active'`, userID)
	if err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read subscription cancellation result: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

var _ Repository = (*PostgresRepository)(nil)
