package commands

import (
	"fmt"

	"github.com/Gasoid/regular-go-bot/v3/messenger"
)

type Command interface {
	Handler(*messenger.Sender, messenger.Receiver) error
	Help() string
	Name() string
}

var (
	commands = []Command{}
)

func Register(command Command) {
	commands = append(commands, command)
}

func List() []Command {
	return commands
}

func Help() string {
	text := ""
	for _, c := range commands {
		text = fmt.Sprintf("%s/%s - %s\n", text, c.Name(), c.Help())
	}
	return text
}

type Callback struct {
	SendMessage  func(text string)
	SendPhoto    func(data, caption string)
	SendVideo    func(filePath string)
	ReplyMessage func(text string)
}
