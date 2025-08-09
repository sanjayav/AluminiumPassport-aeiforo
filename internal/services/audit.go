package services

import (
	"fmt"
	"log"
	"os"
	"time"
)

var auditLog *os.File

func init() {
	var err error
	auditLog, err = os.OpenFile("audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open audit log: %v", err)
	}
}

func LogEvent(username, role, action, resource string) {
	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s] USER=%s ROLE=%s ACTION=%s RESOURCE=%s\n", timestamp, username, role, action, resource)
	_, _ = auditLog.WriteString(entry)
}
