package postgresstorage

import (
	"context"
	"errors"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresStorage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) storage.Storage {
	return &postgresStorage{
		pool: pool,
	}
}

func (ps *postgresStorage) Load() ([]model.URLRecord, error) {
	if ps.pool == nil {
		return nil, errors.New("database connection is nil")
	}

	rows, err := ps.pool.Query(context.Background(),
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

func (ps *postgresStorage) Append(record model.URLRecord) error {
	if ps.pool == nil {
		return errors.New("database connection is nil")
	}

	_, err := ps.pool.Exec(context.Background(),
		"INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		record.UUID, record.ShortURL, record.OriginalURL)

	return err
}

func (ps *postgresStorage) AppendBatch(records []model.URLRecord) error {
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

	br := ps.pool.SendBatch(context.Background(), batch)
	defer br.Close()

	for range records {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}
