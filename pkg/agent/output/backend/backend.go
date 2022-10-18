package backend

import "github.com/sirupsen/logrus"

var (
	log *logrus.Logger
)

// SetLogger adds a logger to this module
func SetLogger(l *logrus.Logger) {
	log = l
}
