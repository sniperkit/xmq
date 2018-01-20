package queue

type Consumer interface {
	GetMessages(count int) ([]Message, error)
	GetStats() (*ConsumerStats, error)
	Close() error
}