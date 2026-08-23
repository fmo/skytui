package notifier

type Noop struct{}

func (Noop) Notify(string, string) error {
	return nil
}
