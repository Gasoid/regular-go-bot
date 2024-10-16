package commands

import "fmt"

type Command interface {
	Handler(string, func(string)) error
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
