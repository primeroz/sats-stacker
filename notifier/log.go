package notifier

import "github.com/sirupsen/logrus"

var log *logrus.Entry

// UseLogger tells the backends package which logger to use
func UseLogger(l *logrus.Logger, name string) {
	log = l.WithFields(logrus.Fields{"notifier": name})
}
