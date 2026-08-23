package notifier

func New() Notifier {
	return newLinux(systemCommandRunner{})
}
