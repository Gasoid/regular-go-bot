package commands

import (
	"fmt"
	"time"

	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/Gasoid/regular-go-bot/v3/metrics"
)

type Command interface {
	Handler(*messenger.Sender, messenger.Receiver) error
	Help() string
	Name() string
}

var (
	commands = []Command{}
	Commands = map[string]Command{}
)

type Wrapper struct {
	command Command
}

func (w *Wrapper) Handler(s *messenger.Sender, r messenger.Receiver) error {
	start := time.Now()
	err := w.command.Handler(s, r)
	metrics.CommandInc(w.command.Name(), err)
	metrics.CommandDuration(w.command.Name(), time.Since(start))
	return err
}

func (w *Wrapper) Help() string {
	return w.command.Help()
}

func (w *Wrapper) Name() string {
	return w.command.Name()
}

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
