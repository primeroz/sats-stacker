package exchange

import "github.com/sirupsen/logrus"
import "strings"

var log *logrus.Entry

// UseLogger tells the backends package which logger to use
func UseLogger(l *logrus.Logger, name string) {
	log = l.WithFields(logrus.Fields{
		"exchange": strings.ToTitle(name),
	})
}
