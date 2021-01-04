package logger

import (
	"fmt"
	"log"
	"os"
)

const (
	logFlags = log.LstdFlags | log.Lshortfile | log.Lmsgprefix
)

func GetLogger(funcName string) *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", funcName), logFlags)
}
