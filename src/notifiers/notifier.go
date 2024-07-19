package notifiers

type Notifier interface {
	SendAlert(message string) error
}
