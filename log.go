package main

import (
	"io"
	"log"
)

type logger struct {
	info    *log.Logger
	warning *log.Logger
	errors  *log.Logger
	debug   *log.Logger
}

func (l *logger) init(infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer, debugHandle io.Writer) {
	l.info = log.New(infoHandle, "INFO   : ", log.Ldate|log.Ltime|log.Lshortfile)
	l.warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	l.errors = log.New(errorHandle, "ERROR  : ", log.Ldate|log.Ltime|log.Lshortfile)
	l.debug = log.New(debugHandle, "DEBUG  : ", log.Ldate|log.Ltime|log.Lshortfile)
}
