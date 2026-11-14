// Package sqlstore stores messages in a SQL database through database/sql.
// It targets SQLite dialects, including Cloudflare D1.
package sqlstore

import (
	"context"
	"crypt-link/core/domain"
	"crypt-link/core/port"
	"database/sql"
	"errors"
	"time"
)

// Schema creates the messages table. There is no index on expires_at: on D1
// every index update counts as a row written, while the hourly purge only
// scans a small table.
const Schema = `CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	encrypted TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL
);`

// MessageRepository is a port.MessageRepository backed by database/sql.
type MessageRepository struct {
	db *sql.DB
}

var _ port.MessageRepository = (*MessageRepository)(nil)

// NewMessageRepository returns a repository using db. The schema must exist.
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(ctx context.Context, msg domain.Message) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO messages (id, encrypted, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		msg.ID, msg.Encrypted, msg.CreatedAt.UnixMilli(), msg.ExpiresAt.UnixMilli())
	return err
}

// FindAndDelete reads and deletes the message in one statement, so a message
// can be returned to at most one caller.
func (r *MessageRepository) FindAndDelete(ctx context.Context, id string) (*domain.Message, error) {
	var (
		msg                  domain.Message
		createdAt, expiresAt int64
	)
	err := r.db.QueryRowContext(ctx,
		`DELETE FROM messages WHERE id = ? RETURNING id, encrypted, created_at, expires_at`, id).
		Scan(&msg.ID, &msg.Encrypted, &createdAt, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	msg.CreatedAt = time.UnixMilli(createdAt).UTC()
	msg.ExpiresAt = time.UnixMilli(expiresAt).UTC()
	return &msg, nil
}

// DeleteExpired removes messages whose expiry is before now and returns how many were removed.
func (r *MessageRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM messages WHERE expires_at < ?`, now.UnixMilli())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
