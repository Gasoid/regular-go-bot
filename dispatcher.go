package main

import (
	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
)

type Dispatcher struct {
	commands    map[string]commands.Command
	middlewares []func(messenger.Handler) messenger.Handler
	eventFuncs  map[messenger.EventType]func(*messenger.Sender, messenger.Receiver) error
}

func (d *Dispatcher) Process(s *messenger.Sender, r messenger.Receiver) error {
	h := processMiddlewares(
		messenger.HandlerFunc(func(sender *messenger.Sender, receiver messenger.Receiver) error {
			switch sender.MessageType {
			case messenger.Command:
				if c, ok := d.commands[sender.Message.Command]; ok {
					return c.Handler(sender, receiver)
				}
			default:
				if f, ok := d.eventFuncs[sender.MessageType]; ok {
					return f(sender, receiver)
				}
			}

			return nil
		}), d.middlewares...)

	err := h.Process(s, r)
	if err != nil {
		return err
	}

	return nil
}
