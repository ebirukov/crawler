package log

import (
	"fmt"
	"time"
)

type Logger interface {
	Log(message string)
}

type Adapter func(message string)

func (l Adapter) Log(message string) {
	l(fmt.Sprintf("%v: %s", time.Now(), message))
}

func Printer(message string) {
	fmt.Println(message)
}
