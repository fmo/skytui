package notifier

func New() Notifier {
	return MacOS{}
}
