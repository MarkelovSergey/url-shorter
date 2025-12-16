package healthservice

import (
	"context"

	"github.com/MarkelovSergey/url-shorter/internal/repository/healthrepository"
)

type HealthService interface {
	Ping(ctx context.Context) error
}

type healthService struct {
	healthRepo healthrepository.HealthRepository
}

func New(healthRepo healthrepository.HealthRepository) HealthService {
	return &healthService{healthRepo}
}

func (s *healthService) Ping(ctx context.Context) error {
	return s.healthRepo.Ping(ctx)
}
