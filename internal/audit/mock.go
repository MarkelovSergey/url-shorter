// Package audit предоставляет функциональность аудита действий пользователей.
package audit

// MockPublisher представляет мок-реализацию издателя для тестирования.
type MockPublisher struct {
	Events []Event
}

// NewMockPublisher создает новый мок-издатель.
func NewMockPublisher() *MockPublisher {
	return &MockPublisher{
		Events: make([]Event, 0),
	}
}

// Subscribe добавляет наблюдателя (заглушка).
func (m *MockPublisher) Subscribe(observer Observer) {}

// Unsubscribe удаляет наблюдателя (заглушка).
func (m *MockPublisher) Unsubscribe(observer Observer) {}

// Publish публикует событие и сохраняет его для проверки.
func (m *MockPublisher) Publish(event Event) {
	m.Events = append(m.Events, event)
}

// Close закрывает издатель (заглушка).
func (m *MockPublisher) Close() error {
	return nil
}
