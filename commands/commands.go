package commands

import (
	"fmt"

	"github.com/Gasoid/regular-go-bot/metrics"
)

type Command interface {
	Handler(string, Callback) error
	Help() string
	Name() string
}

var (
	commands = []Command{}
)

type Wrapper struct {
	command Command
}

func (w *Wrapper) Handler(s string, c Callback) error {
	err := w.command.Handler(s, c)
	metrics.CommandInc(w.command.Name(), err)
	return err
}

func (w *Wrapper) Help() string {
	return w.command.Help()
}

func (w *Wrapper) Name() string {
	return w.command.Name()
}

func Register(command Command) {
	commands = append(commands, &Wrapper{command})
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
	SendVideo    func(filePath string)
	ReplyMessage func(text string)
}
