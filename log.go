package gws

import "go.uber.org/zap"

type Log interface {
	Debug(c *Context, msg string)
	Info(c *Context, msg string, )
	Notice(c *Context, msg string, )
	Warn(c *Context, msg string, )
	Error(c *Context, msg string, )
	Fatal(c *Context, msg string)
}

type Logger struct {
	Logger *zap.Logger
}

func (l *Logger) Debug(c *Context, msg string) {
	l.Logger.Debug(msg)
}
func (l *Logger) Info(c *Context, msg string) {
	l.Logger.Info(msg)
}
func (l *Logger) Notice(c *Context, msg string) {

}
func (l *Logger) Warn(c *Context, msg string) {

}
func (l *Logger) Error(c *Context, msg string) {

}
func (l *Logger) Fatal(c *Context, msg string) {

}
