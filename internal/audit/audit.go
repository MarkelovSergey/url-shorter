package audit

import "time"

type Action string

const (
	ActionShorten Action = "shorten"
	ActionFollow  Action = "follow"
)

type Event struct {
	Timestamp int64   `json:"ts"`      // unix timestamp события
	Action    Action  `json:"action"`  // действие: shorten или follow
	UserID    *string `json:"user_id"` // идентификатор пользователя
	URL       string  `json:"url"`     // оригинальный URL
}

func NewEvent(action Action, url string, userID *string) Event {
	return Event{
		Timestamp: time.Now().Unix(),
		Action:    action,
		URL:       url,
		UserID:    userID,
	}
}

type Observer interface {
	OnEvent(event Event) error
	Close() error
}

type Publisher interface {
	Subscribe(observer Observer)
	Unsubscribe(observer Observer)
	Publish(event Event)
	Close() error
}
