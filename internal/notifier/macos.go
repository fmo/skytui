package notifier

type MacOS struct{}

func (MacOS) Notify(title, message string) error {
	return nil
}
