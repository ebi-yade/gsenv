package main

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

func main() {
	logLevelCandidates := []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"}
	logFilter := &logutils.LevelFilter{
		Levels:   logLevelCandidates,
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stderr,
	}
	logLevel := os.Getenv("GSENV_LOG_LEVEL")
	for _, candidate := range logLevelCandidates {
		if string(candidate) == logLevel {
			log.Println("set log level to", logLevel)
			logFilter.MinLevel = candidate
		}
	}
	log.SetOutput(logFilter)
}
