package notification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"

	"github.com/akhmed9505/delayed-notifier/internal/domain"
)

var ErrNotificationNotFound = errors.New("notification not found")

type Repository struct {
	db *dbpg.DB
}

func New(db *dbpg.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, notification domain.Notification) (uuid.UUID, error) {
	const op = "notification.repository.Create"

	query := `
		INSERT INTO notifications (
			message,
			channel,
			recipient,
			send_at,
			status,
			retries,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id uuid.UUID

	err := r.db.QueryRowContext(
		ctx,
		query,
		notification.Message,
		string(notification.Channel),
		notification.Recipient,
		notification.SendAt,
		string(notification.Status),
		notification.Retries,
		notification.CreatedAt,
		notification.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: create notification: %w", op, err)
	}

	return id, nil
}

func (r *Repository) GetStatusByID(ctx context.Context, id uuid.UUID) (domain.NotificationStatus, error) {
	const op = "notification.repository.GetStatusByID"

	query := `
		SELECT status
		FROM notifications
		WHERE id = $1
	`

	var status domain.NotificationStatus

	err := r.db.QueryRowContext(ctx, query, id).Scan(&status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, ErrNotificationNotFound)
		}

		return "", fmt.Errorf("%s: get notification status by id: %w", op, err)
	}

	return status, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.NotificationStatus) error {
	const op = "notification.repository.UpdateStatus"

	query := `
		UPDATE notifications
		SET status = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	res, err := r.db.ExecContext(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("%s: update notification status: %w", op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: affected rows: %w", op, err)
	}

	if rows == 0 {
		return fmt.Errorf("%s: %w", op, ErrNotificationNotFound)
	}

	return nil
}
