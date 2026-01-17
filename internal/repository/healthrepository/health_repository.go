// Package healthrepository содержит репозиторий для проверки здоровья системы.
package healthrepository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthRepository определяет интерфейс для проверки здоровья.
type HealthRepository interface {
	Ping(ctx context.Context) error
}

type healthRepository struct {
	pool *pgxpool.Pool
}

// New создает новый экземпляр HealthRepository.
func New(conn *pgxpool.Pool) HealthRepository {
	return &healthRepository{conn}
}

// Ping проверяет соединение с базой данных.
func (r *healthRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
