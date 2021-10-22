package utils

import "github.com/sirupsen/logrus"

// Logger interface serve as a generic model to be used with *logrus.Logger or *logrus.Entry
type Logger interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	WithFields(fields logrus.Fields) *logrus.Entry
}
