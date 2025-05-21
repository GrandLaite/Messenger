package notifier

import (
	"fmt"
)

// Notifier — контракт для любого способа уведомления.
type Notifier interface {
	Send(to, subject, body string) error
}

// Builder возвращает готовый объект-уведомитель.
type Builder func() (Notifier, error)

var registry = make(map[string]Builder)

// Register вызывается из init-функций конкретных реализаций.
func Register(kind string, b Builder) {
	registry[kind] = b
}

// Create отдаёт нужный Notifier по имени (email, sms, push …).
func Create(kind string) (Notifier, error) {
	if b, ok := registry[kind]; ok {
		return b()
	}
	return nil, fmt.Errorf("unknown notifier type %q", kind)
}
