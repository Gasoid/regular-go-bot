package help

import (
	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
)

type Command struct{}

func (c *Command) Name() string {
	return "help"
}

func (c *Command) Help() string {
	return "the command prints help"
}

func (c *Command) Args() []string {
	return nil
}

//nolint:errcheck
func (c *Command) Handler(_ *messenger.Sender, r messenger.Receiver) error {
	r.SendMessage(commands.Help())
	return nil
}

func init() {
	commands.Register(&Command{})
}
