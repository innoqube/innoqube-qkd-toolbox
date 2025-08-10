package iqutils

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger
}

func IQLogger(debugEnabled bool) *Logger {
	var debugOutput = io.Discard
	if debugEnabled {
		debugOutput = os.Stdout
	}
	return &Logger{
		debug: log.New(debugOutput, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		warn:  log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
		err:   log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}
