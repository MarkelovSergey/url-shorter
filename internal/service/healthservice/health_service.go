// Package healthservice содержит сервис проверки здоровья системы.
package healthservice

import (
	"context"

	"github.com/MarkelovSergey/url-shorter/internal/repository/healthrepository"
)

// HealthService определяет интерфейс сервиса здоровья.
type HealthService interface {
	Ping(ctx context.Context) error
}

type healthService struct {
	healthRepo healthrepository.HealthRepository
}

// New создает новый экземпляр HealthService.
func New(healthRepo healthrepository.HealthRepository) HealthService {
	return &healthService{healthRepo}
}

// Ping проверяет доступность базы данных.
func (s *healthService) Ping(ctx context.Context) error {
	return s.healthRepo.Ping(ctx)
}
