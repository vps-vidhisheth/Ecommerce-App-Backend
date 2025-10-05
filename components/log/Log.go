package log

import "fmt"

type Logger interface {
	Print(value ...interface{})
	GetLogger()
}

type Log struct {
}

func GetLogger() *Log {
	return &Log{}
}

func (l *Log) Print(value ...interface{}) {
	fmt.Println(value)
	return
}
