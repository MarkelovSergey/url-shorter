package healthrepository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type healthRepository struct {
	pool *pgxpool.Pool
}

func New(conn *pgxpool.Pool) HealthRepository {
	return &healthRepository{conn}
}

func (r *healthRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
