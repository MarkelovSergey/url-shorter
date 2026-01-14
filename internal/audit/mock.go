package audit

type MockPublisher struct {
	Events []Event
}

func NewMockPublisher() *MockPublisher {
	return &MockPublisher{
		Events: make([]Event, 0),
	}
}

func (m *MockPublisher) Subscribe(observer Observer) {}

func (m *MockPublisher) Unsubscribe(observer Observer) {}

func (m *MockPublisher) Publish(event Event) {
	m.Events = append(m.Events, event)
}

func (m *MockPublisher) Close() error {
	return nil
}
