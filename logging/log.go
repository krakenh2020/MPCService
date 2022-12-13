package logging

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func LogSetUp(logLevel, logFile string) {
	switch logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}

	//file, err := os.Create(logFile)
	//if err != nil {
	//	log.Error(err)
	//}
	//defer file.Close()

	//log.SetOutput(io.MultiWriter(file, os.Stderr))
	log.SetOutput(io.MultiWriter(os.Stderr))
}
