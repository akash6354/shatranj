package payments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) CreatePending(ctx context.Context, userID string, amount int64, currency string) (Payment, error) {
	return scanPayment(r.db.QueryRowContext(ctx, `
		INSERT INTO payments (user_id, provider, amount_paise, currency, status)
		VALUES ($1, 'razorpay', $2, $3, 'pending')
		RETURNING id::text, user_id::text, provider, COALESCE(provider_order_id, ''),
			COALESCE(provider_payment_id, ''), amount_paise, currency, status, created_at, updated_at`,
		userID, amount, currency))
}

func (r *PostgresRepository) SetProviderOrder(ctx context.Context, paymentID, orderID string) (Payment, error) {
	payment, err := scanPayment(r.db.QueryRowContext(ctx, `
		UPDATE payments SET provider_order_id = $2, updated_at = now()
		WHERE id = $1 AND status = 'pending'
		RETURNING id::text, user_id::text, provider, COALESCE(provider_order_id, ''),
			COALESCE(provider_payment_id, ''), amount_paise, currency, status, created_at, updated_at`,
		paymentID, orderID))
	if errors.Is(err, sql.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	return payment, err
}

func (r *PostgresRepository) MarkFailed(ctx context.Context, paymentID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE payments SET status = 'failed', updated_at = now()
		WHERE id = $1 AND status = 'pending'`, paymentID)
	if err != nil {
		return fmt.Errorf("mark payment failed: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Capture(ctx context.Context, orderID, paymentID string, amount int64, currency string) (Payment, error) {
	payment, err := scanPayment(r.db.QueryRowContext(ctx, `
		UPDATE payments SET status = 'paid', provider_payment_id = NULLIF($2, ''), updated_at = now()
		WHERE provider_order_id = $1 AND amount_paise = $3 AND currency = $4 AND status = 'pending'
		RETURNING id::text, user_id::text, provider, COALESCE(provider_order_id, ''),
			COALESCE(provider_payment_id, ''), amount_paise, currency, status, created_at, updated_at`,
		orderID, paymentID, amount, currency))
	if err == nil {
		return payment, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Payment{}, err
	}
	payment, err = scanPayment(r.db.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, provider, COALESCE(provider_order_id, ''),
			COALESCE(provider_payment_id, ''), amount_paise, currency, status, created_at, updated_at
		FROM payments WHERE provider_order_id = $1`, orderID))
	if errors.Is(err, sql.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	if err != nil {
		return Payment{}, err
	}
	if payment.AmountPaise != amount || payment.Currency != currency || payment.Status != "paid" ||
		(paymentID != "" && payment.ProviderPaymentID != "" && payment.ProviderPaymentID != paymentID) {
		return Payment{}, ErrInvalid
	}
	if paymentID != "" && payment.ProviderPaymentID == "" {
		if _, err := r.db.ExecContext(ctx, `
			UPDATE payments SET provider_payment_id = $2, updated_at = now()
			WHERE id = $1 AND status = 'paid' AND provider_payment_id IS NULL`, payment.ID, paymentID); err != nil {
			return Payment{}, fmt.Errorf("save captured payment ID: %w", err)
		}
		payment.ProviderPaymentID = paymentID
	}
	return payment, nil
}

func (r *PostgresRepository) List(ctx context.Context, userID string, limit, offset int) ([]Payment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, provider, COALESCE(provider_order_id, ''),
			COALESCE(provider_payment_id, ''), amount_paise, currency, status, created_at, updated_at
		FROM payments WHERE user_id = $1 ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()
	items := make([]Payment, 0)
	for rows.Next() {
		item, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payments: %w", err)
	}
	return items, nil
}

type scanner interface{ Scan(...any) error }

func scanPayment(row scanner) (Payment, error) {
	var item Payment
	err := row.Scan(&item.ID, &item.UserID, &item.Provider, &item.ProviderOrderID,
		&item.ProviderPaymentID, &item.AmountPaise, &item.Currency, &item.Status,
		&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Payment{}, fmt.Errorf("read payment: %w", err)
	}
	return item, nil
}

var _ Repository = (*PostgresRepository)(nil)
