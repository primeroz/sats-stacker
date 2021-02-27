package notifier

import (
	//"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"strings"
)

type Stdout struct {
	Name string
}

//Config the Stdout object

func (s *Stdout) Config(c *cli.Context) error {
	s.Name = strings.ToTitle("stdout")

	return nil
}

func (s *Stdout) Notify(title string, message string) (e error) {

	log.WithFields(logrus.Fields{
		"action": "notify",
	}).Debug("Notifying using " + s.Name)

	fmt.Printf("\n%s\n\n", title)
	fmt.Printf("%s", message)

	return nil
}
