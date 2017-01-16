package config

import (
	"github.com/Sirupsen/logrus"
)

var (
	//Log the Logger
	log     *logrus.Logger
	dataDir string
	logDir  string
	confDir string
)

// SetDirs set default dirs to set db and logs
func SetDirs(data string, log string, conf string) {
	dataDir = data
	logDir = log
	confDir = conf
}

// SetLogger set the output log
func SetLogger(l *logrus.Logger) {
	log = l
}
