package notifier

func New() Notifier {
	return Noop{}
}
