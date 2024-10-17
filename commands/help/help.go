package help

import (
	"github.com/Gasoid/regular-go-bot/commands"
)

type Command struct{}

func (c *Command) Name() string {
	return "help"
}

func (c *Command) Help() string {
	return "the command prints help"
}

func (c *Command) Handler(s string, callback func(string)) error {
	callback(commands.Help())
	return nil
}

func init() {
	commands.Register(&Command{})
}
