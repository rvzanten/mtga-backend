package types

import (
	"io"
	"log"
)

// Logger logs
type Logger struct {
	Info    *log.Logger
	Warning *log.Logger
	Errors  *log.Logger
	Debug   *log.Logger
}

func (l *Logger) Init(infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer, debugHandle io.Writer) {
	l.Info = log.New(infoHandle, "INFO   : ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Errors = log.New(errorHandle, "ERROR  : ", log.Ldate|log.Ltime|log.Lshortfile)
	l.Debug = log.New(debugHandle, "DEBUG  : ", log.Ldate|log.Ltime|log.Lshortfile)
}
