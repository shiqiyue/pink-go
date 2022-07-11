package log

import (
	"context"
	"fmt"
	"log"
)

//系统日志
type Logger interface {
	Info(ctx context.Context, format string, a ...interface{})
	Error(ctx context.Context, format string, a ...interface{})
}

type StdLogger struct {
}

func (l *StdLogger) Info(ctx context.Context, format string, a ...interface{}) {
	log.Println(fmt.Sprintf(format, a...))
}

func (l *StdLogger) Error(ctx context.Context, format string, a ...interface{}) {
	log.Println(fmt.Sprintf(format, a...))
}
