package healthrepository

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type healthRepository struct {
	conn *pgx.Conn
}

func New(conn *pgx.Conn) HealthRepository {
	return &healthRepository{conn}
}

func (r *healthRepository) Ping(ctx context.Context) error {
	return r.conn.Ping(ctx)
}
