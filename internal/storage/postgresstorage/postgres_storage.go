// Package postgresstorage содержит реализацию хранилища на основе PostgreSQL.
package postgresstorage

import (
	"context"
	"errors"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresStorage представляет PostgreSQL-хранилище.
type PostgresStorage struct {
	pool *pgxpool.Pool
}

// New создает новое PostgreSQL-хранилище.
func New(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool}
}

// Load загружает все записи из базы данных.
func (ps *PostgresStorage) Load(ctx context.Context) ([]model.URLRecord, error) {
	rows, err := ps.pool.Query(ctx,
		"SELECT uuid, short_url, original_url, COALESCE(user_id, ''), COALESCE(is_deleted, false) FROM urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.URLRecord
	for rows.Next() {
		var record model.URLRecord
		if err := rows.Scan(&record.UUID, &record.ShortURL, &record.OriginalURL, &record.UserID, &record.IsDeleted); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// Append добавляет запись в базу данных.
func (ps *PostgresStorage) Append(ctx context.Context, record model.URLRecord) error {
	_, err := ps.pool.Exec(ctx,
		"INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)",
		record.UUID, record.ShortURL, record.OriginalURL, record.UserID)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return repository.ErrURLAlreadyExists
	}

	return err
}

// AppendBatch добавляет несколько записей в базу данных.
func (ps *PostgresStorage) AppendBatch(ctx context.Context, records []model.URLRecord) error {
	if len(records) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, record := range records {
		batch.Queue(
			"INSERT INTO urls (uuid, short_url, original_url, user_id) VALUES ($1, $2, $3, $4)",
			record.UUID, record.ShortURL, record.OriginalURL, record.UserID,
		)
	}

	br := ps.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range records {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

// FindByOriginalURL находит короткий URL по оригинальному.
func (ps *PostgresStorage) FindByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	var shortURL string
	err := ps.pool.QueryRow(
		ctx,
		"SELECT short_url FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&shortURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrNotFound
		}

		return "", err
	}

	return shortURL, nil
}

// FindByShortURL находит оригинальный URL по короткому.
func (ps *PostgresStorage) FindByShortURL(ctx context.Context, shortURL string) (string, error) {
	var (
		originalURL string
		isDeleted   bool
	)

	err := ps.pool.QueryRow(
		ctx,
		"SELECT original_url, COALESCE(is_deleted, false) FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&originalURL, &isDeleted)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrNotFound
		}

		return "", err
	}

	if isDeleted {
		return "", repository.ErrDeleted
	}

	return originalURL, nil
}

// FindByUserID находит все URL пользователя.
func (ps *PostgresStorage) FindByUserID(ctx context.Context, userID string) ([]model.URLRecord, error) {
	rows, err := ps.pool.Query(ctx,
		"SELECT uuid, short_url, original_url, COALESCE(user_id, ''), COALESCE(is_deleted, false) FROM urls WHERE user_id = $1",
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.URLRecord
	for rows.Next() {
		var record model.URLRecord
		if err := rows.Scan(&record.UUID, &record.ShortURL, &record.OriginalURL, &record.UserID, &record.IsDeleted); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// DeleteBatch удаляет несколько URL пакетно.
func (ps *PostgresStorage) DeleteBatch(ctx context.Context, shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, shortURL := range shortURLs {
		batch.Queue(
			"UPDATE urls SET is_deleted = true WHERE short_url = $1 AND user_id = $2",
			shortURL, userID,
		)
	}

	br := ps.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range shortURLs {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}
