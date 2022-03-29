package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotContext struct {
	Update *tgbotapi.Update
	Msg    *tgbotapi.MessageConfig
	Action tgbotapi.Chattable
}

type Bot struct {
	api      *tgbotapi.BotAPI
	Context  *BotContext
	isOzon   bool
	commands map[string]func(c *BotContext)
}

func New() *Bot {
	api, err := tgbotapi.NewBotAPI(os.Getenv(token))
	if err != nil {
		log.Fatal("unexpected error:", err)
	}
	api.Debug = true
	log.Printf("Authorized on account %s", api.Self.UserName)
	bot := &Bot{
		api: api,
	}
	if os.Getenv(enableOzon) == "true" {
		bot.isOzon = true
	}
	bot.commands = make(map[string]func(c *BotContext))
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
	return &BotContext{
		Update: update,
		Msg:    &msg,
	}
}

func (b *Bot) Flush(c *BotContext) {
	if c.Msg.Text != "" {
		if _, err := b.api.Send(c.Msg); err != nil {
			log.Println(err.Error())
		}
	}
	if c.Action != nil {
		if _, err := b.api.Send(c.Action); err != nil {
			log.Println(err.Error())
		}
	}
}

func (b *Bot) Command(cmd string, f func(c *BotContext)) {
	b.commands[cmd] = f
}

func (b *Bot) HandleCommand(update *tgbotapi.Update) {
	if !update.Message.IsCommand() {
		return
	}
	c := b.NewBotContext(update)
	if f, ok := b.commands[update.Message.Command()]; ok {
		f(c)
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

func (c *BotContext) Text(text string) {
	c.Msg.Text = text
}
