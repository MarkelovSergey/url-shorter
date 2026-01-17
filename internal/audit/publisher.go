package audit

import (
	"slices"
	"sync"

	"go.uber.org/zap"
)

// AuditPublisher реализует паттерн Publisher для рассылки событий аудита.
type AuditPublisher struct {
	observers []Observer
	mu        *sync.RWMutex
	logger    *zap.Logger
}

// NewPublisher создает новый экземпляр AuditPublisher.
func NewPublisher(logger *zap.Logger) *AuditPublisher {
	return &AuditPublisher{
		observers: make([]Observer, 0),
		mu:        &sync.RWMutex{},
		logger:    logger,
	}
}

// Subscribe добавляет наблюдателя для получения событий аудита.
func (p *AuditPublisher) Subscribe(observer Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.observers = append(p.observers, observer)
	p.logger.Info("audit observer subscribed", zap.Int("total_observers", len(p.observers)))
}

// Unsubscribe удаляет наблюдателя из списка подписчиков.
func (p *AuditPublisher) Unsubscribe(observer Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, obs := range p.observers {
		if obs == observer {
			p.observers = slices.Delete(p.observers, i, i+1)
			p.logger.Info("audit observer unsubscribed", zap.Int("total_observers", len(p.observers)))
			return
		}
	}
}

// Publish рассылает событие всем наблюдателям асинхронно.
func (p *AuditPublisher) Publish(event Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, observer := range p.observers {
		go func(obs Observer) {
			if err := obs.OnEvent(event); err != nil {
				p.logger.Error("failed to send audit event to observer",
					zap.Error(err),
					zap.String("action", string(event.Action)),
					zap.String("url", event.URL),
				)
			}
		}(observer)
	}
}

// Close закрывает всех наблюдателей и освобождает ресурсы.
func (p *AuditPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, observer := range p.observers {
		if err := observer.Close(); err != nil {
			p.logger.Error("failed to close audit observer", zap.Error(err))
		}
	}

	p.observers = nil
	p.logger.Info("audit publisher closed")

	return nil
}

// HasObservers проверяет, есть ли зарегистрированные наблюдатели.
func (p *AuditPublisher) HasObservers() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.observers) > 0
}
