package notifier

import "github.com/urfave/cli/v2"

type Notifier interface {
	Config(c *cli.Context) (err error)
	Notify(title string, message string) (err error)
}
