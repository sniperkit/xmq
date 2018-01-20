package nsq

import (
	nsq "github.com/bitly/go-nsq"
)

type Config struct {
	Addr               string
	Topic              string
	Channel            string
	Concurrency        uint
	ConsumerBufferSize uint
	LogLevel           LogLevel
	Nsq                *nsq.Config
}
