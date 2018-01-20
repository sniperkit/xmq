package queue


type MessagePayload struct {
	Id   string                 `json:"id"`
	Hash string                 `json:"hash"`
	Body map[string]interface{} `json:"body"`
}


type Message interface {
	Acknowledge() error
	Reject() error
	Payload() *MessagePayload
}

