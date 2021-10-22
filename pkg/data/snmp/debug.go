package snmp

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

type writer struct {
	io.Writer
	timeFormat string
}

func (w writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)), b...))
}

// GetDebugLogger returns a logger handler for snmp debug data
func GetDebugLogger(filename string) gosnmp.Logger {
	name := filepath.Join(logDir, "snmpdebug_"+strings.Replace(filename, ".", "-", -1)+".log")
	l, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		mainlog.Warnf("Error on create debug file : %s ", err)
		return gosnmp.NewLogger(log.New(os.Stdout, "", 0))
	}
	return gosnmp.NewLogger(log.New(&writer{l, "2006-01-02 15:04:05.00000"}, " [SNMP-DEBUG] ", 0))
}

//"github.com/gosnmp/gosnmp"
