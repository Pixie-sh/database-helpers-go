package database

import (
	"github.com/pixie-sh/logger-go/logger"
)

type log struct {
	logger.Interface
}

func (l log) Printf(log string, args ...any) {
	l.Log(log, args...)
}
