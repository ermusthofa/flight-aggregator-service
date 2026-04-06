package pkg

import (
	"log"
)

func Info(msg string, args ...interface{}) {
	log.Printf("[INFO] "+msg, args...)
}

func Warning(msg string, args ...interface{}) {
	log.Printf("[WARNING] "+msg, args...)
}

func Error(msg string, args ...interface{}) {
	log.Printf("[ERROR] "+msg, args...)
}
