package pomodoro

type fakeNotifier struct {
	calls   int
	title   string
	message string
	err     error
}

func (f *fakeNotifier) Notify(title, message string) error {
	f.calls++
	f.title = title
	f.message = message
	return f.err
}
