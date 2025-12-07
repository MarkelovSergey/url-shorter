package postgresstorage

import (
	"context"
	"errors"

	"github.com/MarkelovSergey/url-shorter/internal/model"
	"github.com/MarkelovSergey/url-shorter/internal/storage"
	"github.com/jackc/pgx/v5"
)

type postgresStorage struct {
	conn *pgx.Conn
}

func New(conn *pgx.Conn) storage.Storage {
	return &postgresStorage{
		conn: conn,
	}
}

func (ps *postgresStorage) Load() ([]model.URLRecord, error) {
	if ps.conn == nil {
		return nil, errors.New("database connection is nil")
	}

	rows, err := ps.conn.Query(context.Background(),
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
	if ps.conn == nil {
		return errors.New("database connection is nil")
	}

	_, err := ps.conn.Exec(context.Background(),
		"INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		record.UUID, record.ShortURL, record.OriginalURL)

	return err
}
