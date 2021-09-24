package snmp

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type writer struct {
	io.Writer
	timeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)), b...))
}

// GetDebugLogger returns a logger handler for snmp debug data
func GetDebugLogger(filename string) *log.Logger {
	name := filepath.Join(logDir, "snmpdebug_"+strings.Replace(filename, ".", "-", -1)+".log")
	l, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err == nil {
		return log.New(&writer{l, "2006-01-02 15:04:05.00000"}, " [SNMP-DEBUG] ", 0)
	}
	mainlog.Warnf("Error on create debug file : %s ", err)
	return nil
}
