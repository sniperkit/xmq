package queue

type ConsumerStats struct {
	// Number of messages in queue
	Queue    int `json:"queue"`
	// Number of messages, which has not been acknowledged by consumer
	Unacked  int `json:"unacked"`
	// Number of messages, rejected by consumer
	Rejected int `json:"rejected"`
}
