package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogSetupAndDestruct ..
// (c) https://stackoverflow.com/questions/25342806/proper-pattern-to-encapsulate-log-setup-in-golang
func LogSetupAndDestruct(fpath string) func() {

	if argReportPath == "" {
		// default
		return func() {}
	}

	logFile, err := os.OpenFile(fpath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	if err != nil {
		log.Panicln(err)
	}
	log.SetFlags(0)
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))

	return func() {
		e := logFile.Close()
		if e != nil {
			fmt.Fprintf(os.Stderr, "Problem closing the log file: %s\n", e)
		}
	}
}
