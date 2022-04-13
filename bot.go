package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/prometheus/client_golang/prometheus"
)

type BotContext struct {
	Update        *tgbotapi.Update
	Msg           *tgbotapi.MessageConfig
	Action        tgbotapi.Chattable
	notifications chan *BotContext
}

type Bot struct {
	api           *tgbotapi.BotAPI
	Context       *BotContext
	isOzon        bool
	commands      map[string]func(c *BotContext)
	notifications chan *BotContext
}

func New() *Bot {
	api, err := tgbotapi.NewBotAPI(os.Getenv(token))
	if err != nil {
		log.Fatal("unexpected error:", err)
	}
	// Disable logging
	//api.Debug = true
	log.Printf("Authorized on account %s", api.Self.UserName)
	bot := &Bot{
		api: api,
	}
	if os.Getenv(enableOzon) == "true" {
		bot.isOzon = true
	}
	bot.commands = make(map[string]func(c *BotContext))
	bot.notifications = make(chan *BotContext, 1)
	return bot
}

func (b *Bot) Updates() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout
	return b.api.GetUpdatesChan(u)
}

func (b *Bot) IsOzon(update *tgbotapi.Update) bool {
	return update.Message.Chat.Title == gozone
	// return b.isOzon
}

func (b *Bot) NewBotContext(update *tgbotapi.Update) *BotContext {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	return &BotContext{
		Update:        update,
		Msg:           &msg,
		notifications: b.notifications,
	}
}

func (b *Bot) Flush(c *BotContext) {
	if c.Msg.Text != "" {
		if _, err := b.api.Send(c.Msg); err != nil {
			log.Println(err.Error())
			c.Msg.ParseMode = ""
			log.Println("trying to send it without markdown")
			b.api.Send(c.Msg)
		}
	}
	if c.Action != nil {
		if _, err := b.api.Send(c.Action); err != nil {
			log.Println(err.Error())
		}
	}
}

func (b *Bot) Command(cmd string, f func(c *BotContext), help string) {
	b.commands[cmd] = f
	if help != "" {
		helps = append(helps, fmt.Sprintf("- /%s: %s", cmd, help))
	}
}

func (b *Bot) HandleCommand(update *tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Bot got unexpected error")
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()
	if !update.Message.IsCommand() {
		return
	}
	c := b.NewBotContext(update)
	if f, ok := b.commands[update.Message.Command()]; ok {
		f(c)
		commandExecutions.With(prometheus.Labels{"command": update.Message.Command()}).Inc()
	} else {
		notFound(c)
	}
	b.Flush(c)
}

func (b *Bot) HandleOzon(update *tgbotapi.Update) {
	if b.IsOzon(update) {
		c := b.NewBotContext(update)
		checkOzon(c)
		b.Flush(c)
	}
}

func (b *Bot) HandleNotifications() {
	for c := range b.notifications {
		b.Flush(c)
	}
}

// func (b *Bot) listCommands() []string {
// 	commands := make([]string, len(b.commands))
// 	for c := range b.commands {
// 		commands = append(commands, c)
// 	}
// 	return commands
// }

func (c *BotContext) Text(text string, args ...interface{}) {
	if args != nil {
		c.Msg.Text = fmt.Sprintf(text, args...)
	} else {
		c.Msg.Text = text
	}
	c.Msg.Text = strings.TrimSpace(c.Msg.Text)
}

func (c *BotContext) Notify() {
	c.notifications <- c
}
