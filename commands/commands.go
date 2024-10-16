package commands

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
