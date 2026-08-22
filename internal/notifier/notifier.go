package notifier

type Notifier interface {
	Notify(title, message string) error
}
