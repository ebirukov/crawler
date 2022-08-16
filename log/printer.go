package log

import "fmt"

type Logger interface {
	Log(message string)
}

type Adapter func(message string)

func (l Adapter) Log(message string) {
	l(message)
}

func Printer(message string) {
	fmt.Println(message)
}
