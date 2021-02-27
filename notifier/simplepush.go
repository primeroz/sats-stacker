package notifier

import (
	"github.com/simplepush/simplepush-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"strings"
)

type SimplePush struct {
	Name     string
	Encrypt  bool
	Key      string
	Password string
	Salt     string
	Event    string
}

//Config the Simplepush object

func (s *SimplePush) Config(c *cli.Context) error {
	s.Name = strings.ToTitle("simplepush")
	s.Encrypt = c.Bool("sp-encrypt")
	s.Key = c.String("sp-key")
	s.Password = c.String("sp-password")
	s.Salt = c.String("sp-salt")
	s.Event = c.String("sp-event")

	return nil
}

func (s *SimplePush) Notify(title string, message string) (e error) {

	log.WithFields(logrus.Fields{
		"action": "notify",
	}).Debug("Notifying using " + s.Name)

	msg := simplepush.Message{
		SimplePushKey: s.Key,
		Password:      s.Password,
		Title:         title,
		Message:       message,
		Event:         s.Event,
		Encrypt:       s.Encrypt,
		Salt:          s.Salt,
	}

	err := simplepush.Send(msg)

	if err != nil {
		return err
	}

	return nil
}
