// Package audit предоставляет функциональность аудита действий пользователей.
package audit

import "time"

// Action представляет тип действия для аудита.
type Action string

// Константы действий аудита.
const (
	// ActionShorten - действие сокращения URL.
	ActionShorten Action = "shorten"
	// ActionFollow - действие перехода по короткой ссылке.
	ActionFollow Action = "follow"
)

// Event представляет событие аудита.
type Event struct {
	Timestamp int64   `json:"ts"`      // unix timestamp события
	Action    Action  `json:"action"`  // действие: shorten или follow
	UserID    *string `json:"user_id"` // идентификатор пользователя
	URL       string  `json:"url"`     // оригинальный URL
}

// NewEvent создает новое событие аудита.
func NewEvent(action Action, url string, userID *string) Event {
	return Event{
		Timestamp: time.Now().Unix(),
		Action:    action,
		URL:       url,
		UserID:    userID,
	}
}

// Observer представляет интерфейс наблюдателя событий аудита.
type Observer interface {
	OnEvent(event Event) error
	Close() error
}

// Publisher представляет интерфейс издателя событий аудита.
type Publisher interface {
	Subscribe(observer Observer)
	Unsubscribe(observer Observer)
	Publish(event Event)
	Close() error
}
