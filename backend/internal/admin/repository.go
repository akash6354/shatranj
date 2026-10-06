package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) IsAdmin(ctx context.Context, userID string) (bool, error) {
	var ok bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'admin' AND status = 'active')`,
		userID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check administrator role: %w", err)
	}
	return ok, nil
}

func (r *PostgresRepository) Dashboard(ctx context.Context) (Dashboard, error) {
	var result Dashboard
	err := r.db.QueryRowContext(ctx, `
		SELECT
			(SELECT count(*) FROM users WHERE status = 'active'),
			(SELECT count(*) FROM users WHERE status = 'suspended'),
			(SELECT count(*) FROM payments WHERE status = 'pending'),
			(SELECT count(*) FROM payments WHERE status = 'paid'),
			(SELECT count(*) FROM news_posts WHERE status = 'draft'),
			(SELECT count(*) FROM coach_profiles WHERE status = 'pending'),
			(SELECT count(*) FROM coach_bookings WHERE status = 'pending'),
			(SELECT count(*) FROM admin_payment_flags WHERE status = 'open')`,
	).Scan(&result.ActiveUsers, &result.SuspendedUsers, &result.PendingPayments,
		&result.PaidPayments, &result.DraftNews, &result.PendingCoachApps, &result.PendingBookings,
		&result.OpenPaymentFlags)
	if err != nil {
		return Dashboard{}, fmt.Errorf("load admin dashboard: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) ListUsers(ctx context.Context, query string, limit, offset int) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id::text, u.email, COALESCE(p.username, ''), u.role, u.status, u.created_at
		FROM users u LEFT JOIN profiles p ON p.user_id = u.id
		WHERE $1 = '' OR u.email ILIKE '%' || $1 || '%' OR p.username ILIKE '%' || $1 || '%'
		ORDER BY u.created_at DESC, u.id LIMIT $2 OFFSET $3`, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list admin users: %w", err)
	}
	defer rows.Close()
	items := make([]User, 0)
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.Email, &item.Username, &item.Role, &item.Status, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan admin user: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin users: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) SetUserStatus(ctx context.Context, userID, status string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE users SET status = $2, updated_at = now() WHERE id = $1`, userID, status)
	if err != nil {
		return fmt.Errorf("moderate user status: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read user moderation result: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ListPayments(ctx context.Context, limit, offset int) ([]Payment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pay.id::text, pay.user_id::text, COALESCE(p.username, ''),
			pay.amount_paise, pay.currency, pay.status, COALESCE(pay.provider_order_id, ''), pay.created_at
		FROM payments pay LEFT JOIN profiles p ON p.user_id = pay.user_id
		ORDER BY pay.created_at DESC, pay.id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list admin payments: %w", err)
	}
	defer rows.Close()
	items := make([]Payment, 0)
	for rows.Next() {
		var item Payment
		if err := rows.Scan(&item.ID, &item.UserID, &item.Username, &item.AmountPaise, &item.Currency,
			&item.Status, &item.ProviderOrderID, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan admin payment: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin payments: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) FlagPayment(ctx context.Context, paymentID, adminID, reason string) (PaymentFlag, error) {
	var flag PaymentFlag
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO admin_payment_flags (payment_id, admin_id, reason)
		VALUES ($1, $2, $3)
		RETURNING id::text, payment_id::text, admin_id::text, reason, status, created_at`,
		paymentID, adminID, reason).
		Scan(&flag.ID, &flag.PaymentID, &flag.AdminID, &flag.Reason, &flag.Status, &flag.CreatedAt)
	if err != nil {
		if isForeignKeyViolation(err) {
			return PaymentFlag{}, ErrNotFound
		}
		return PaymentFlag{}, fmt.Errorf("flag payment for review: %w", err)
	}
	return flag, nil
}

func (r *PostgresRepository) ListPaymentFlags(ctx context.Context, limit, offset int) ([]PaymentFlag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, payment_id::text, admin_id::text, reason, status, created_at
		FROM admin_payment_flags WHERE status = 'open'
		ORDER BY created_at, id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list open payment flags: %w", err)
	}
	defer rows.Close()
	items := make([]PaymentFlag, 0)
	for rows.Next() {
		var item PaymentFlag
		if err := rows.Scan(&item.ID, &item.PaymentID, &item.AdminID, &item.Reason, &item.Status, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan payment flag: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate payment flags: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ResolvePaymentFlag(ctx context.Context, flagID, adminID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE admin_payment_flags SET status = 'resolved', resolved_by = $2, resolved_at = now()
		WHERE id = $1 AND status = 'open'`, flagID, adminID)
	if err != nil {
		return fmt.Errorf("resolve payment flag: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read payment flag resolution result: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) SetCoachStatus(ctx context.Context, userID, status string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE coach_profiles SET status = $2, updated_at = now() WHERE user_id = $1`, userID, status)
	if err != nil {
		return fmt.Errorf("moderate coach profile: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read coach moderation result: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func isForeignKeyViolation(err error) bool {
	var pgError interface{ SQLState() string }
	return errors.As(err, &pgError) && pgError.SQLState() == "23503"
}

var _ Repository = (*PostgresRepository)(nil)
