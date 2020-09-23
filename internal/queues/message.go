package queues

// MessageQueue is
type MessageQueue interface {
	Dequeue(string) (string, error)
}
