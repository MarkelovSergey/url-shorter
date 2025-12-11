package postgresstorage

import (
	"context"
	"errors"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/repository"
	"github.com/MarkelovSergey/url-shorter/internal/storage"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresStorage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) storage.Storage {
	return &postgresStorage{pool}
}

func (ps *postgresStorage) Load(ctx context.Context) ([]model.URLRecord, error) {
	if ps.pool == nil {
		return nil, errors.New("database connection is nil")
	}

	rows, err := ps.pool.Query(ctx,
		"SELECT uuid, short_url, original_url FROM urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.URLRecord
	for rows.Next() {
		var record model.URLRecord
		if err := rows.Scan(&record.UUID, &record.ShortURL, &record.OriginalURL); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

func (ps *postgresStorage) Append(ctx context.Context, record model.URLRecord) error {
	if ps.pool == nil {
		return errors.New("database connection is nil")
	}

	_, err := ps.pool.Exec(ctx,
		"INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		record.UUID, record.ShortURL, record.OriginalURL)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return repository.ErrURLAlreadyExists
	}

	return err
}

func (ps *postgresStorage) AppendBatch(ctx context.Context, records []model.URLRecord) error {
	if ps.pool == nil {
		return errors.New("database connection is nil")
	}

	if len(records) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, record := range records {
		batch.Queue(
			"INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)",
			record.UUID, record.ShortURL, record.OriginalURL,
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

func (ps *postgresStorage) FindByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	if ps.pool == nil {
		return "", errors.New("database connection is nil")
	}

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

func (ps *postgresStorage) FindByShortURL(ctx context.Context, shortURL string) (string, error) {
	if ps.pool == nil {
		return "", errors.New("database connection is nil")
	}

	var originalURL string
	err := ps.pool.QueryRow(
		ctx,
		"SELECT original_url FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrNotFound
		}

		return "", err
	}

	return originalURL, nil
}
