package notifier

func New() Notifier {
	return newWindows(systemCommandRunner{})
}
