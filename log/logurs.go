package log

import (
	"errors"
	"io"

	"github.com/Sirupsen/logrus"
)

type Logrus struct {
	logger *logrus.Logger
}

func NewLogrus(writer io.Writer, level Level, format Format) (Interface, error) {
	// Validate level
	logrusLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		return nil, errors.New("unknown log level, choose one from [debug, info, warning, error, fatal, panic]")
	}

	// Validate format
	var logrusFormat logrus.Formatter
	switch format {
	case "text":
		logrusFormat = new(logrus.TextFormatter)
	case "json":
		logrusFormat = new(logrus.JSONFormatter)
	default:
		return nil, errors.New("unknown log format, choose one fromt [text, json]")
	}

	// Create logger
	l := &Logrus{
		logger: logrus.New(),
	}
	l.logger.Out = writer
	l.logger.Formatter = logrusFormat
	l.logger.Hooks = make(logrus.LevelHooks)
	l.logger.Level = logrusLevel

	return l, nil
}

func (l *Logrus) Info(f Fields) {
	l.logger.WithFields(logrus.Fields(f))
}

func (l *Logrus) Warn(f Fields) {
	l.logger.WithFields(logrus.Fields(f))
}

func (l *Logrus) Error(f Fields) {
	l.logger.WithFields(logrus.Fields(f))
}

func (l *Logrus) Fatal(f Fields) {
	l.logger.WithFields(logrus.Fields(f))
}

func (l *Logrus) Panic(f Fields) {
	l.logger.WithFields(logrus.Fields(f))
}