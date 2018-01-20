package log

type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warning"
	Error Level = "error"
	Fatal Level = "fatal"
	Panic Level = "panic"
)

type Format string

const (
	Text Format = "text"
	Json Format = "json"
)

type Fields map[string]interface{}

type Interface interface {
	Info(Fields)
	Warn(Fields)
	Error(Fields)
	Fatal(Fields)
	Panic(Fields)
}

