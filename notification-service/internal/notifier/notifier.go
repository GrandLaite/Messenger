package notifier

type Notifier interface {
	Send(to, subject, body string) error
}
