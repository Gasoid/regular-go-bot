package bot

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const (
	CommandUpdateType UpdateType = "command"
	MessageUpdateType UpdateType = "message"
	UnknownUpdateType UpdateType = "unknown"
)

type Help struct {
	commandList []string
}

func (h *Help) Add(cmd, text string) {
	h.commandList = append(h.commandList, fmt.Sprintf("- /%s: %s", cmd, text))
}

func (h *Help) List() []string {
	return h.commandList
}

func NewHelp() *Help {
	return &Help{commandList: []string{}}
}

type UpdateType string

type MetricsExporter interface {
	CommandInc(string)
}

type Update interface {
	Type() UpdateType
	Command() string
	Args() []string
	Arg() string
	Message() string
	Answer(text string, isReply bool, messageID int)
	User() *User
	ChatID() int64
	ChatTitle() string
	MessageID() int
	NewMembers() []User
}

type User struct {
	ID        int64
	Username  string
	FirstName string
	LastName  string
}

type BotContext struct {
	update         Update
	answer         string
	User           *User
	ChatID         int64
	ChatTitle      string
	Message        string
	reply          bool
	NewChatMembers []User
	answerFunc     func(b *Bot)
}

type Messenger interface {
	Updates() Update
	BanUser(chatID, userID int64)
}

type Bot struct {
	checkers     []func(c *BotContext)
	commands     map[string]func(c *BotContext)
	messenger    Messenger
	newMembersID map[int64]int
	help         *Help
}

func New(messenger Messenger, help *Help, newMembersID map[int64]int) *Bot {
	bot := &Bot{
		messenger:    messenger,
		commands:     make(map[string]func(c *BotContext)),
		help:         help,
		newMembersID: newMembersID,
		checkers:     []func(c *BotContext){},
	}
	return bot
}

func (b *Bot) Run(metrics MetricsExporter) {
	for {
		update := b.messenger.Updates()
		c := &BotContext{
			update:         update,
			User:           update.User(),
			ChatID:         update.ChatID(),
			ChatTitle:      update.ChatTitle(),
			Message:        update.Message(),
			NewChatMembers: update.NewMembers(),
		}
		switch update.Type() {
		case CommandUpdateType:
			b.HandleCommand(c)
			metrics.CommandInc(c.update.Command())
		case MessageUpdateType:
			b.HandleMessage(c)
		}
		if c.answerFunc != nil {
			go c.answerFunc(b)
		}
		b.Flush(c)
	}
}

func (b *Bot) Flush(c *BotContext) {
	if c.answer != "" {
		c.update.Answer(c.answer, c.reply, c.update.MessageID())
		c.answer = ""
	}
}

func (b *Bot) BanUser(chatID int64, userID int64) {
	b.messenger.BanUser(chatID, userID)
}

func (b *Bot) Command(cmd string, f func(c *BotContext), help string) {
	b.commands[cmd] = f
	if help != "" {
		b.help.Add(cmd, help)
	}
}

func (b *Bot) MessageHandler(f func(c *BotContext)) {
	b.checkers = append(b.checkers, f)
}

func (b *Bot) HandleCommand(c *BotContext) {
	defer doRecover()

	if f, ok := b.commands[c.update.Command()]; ok {
		f(c)
	} else {
		notFound(c)
	}
}

func (b *Bot) HandleMessage(c *BotContext) {
	defer doRecover()

	for _, f := range b.checkers {
		f(c)
	}
}

func (b *Bot) HandleNotifications() {

}

func (c *BotContext) Text(text string, args ...interface{}) {
	if args != nil {
		c.answer = fmt.Sprintf(text, args...)
	} else {
		c.answer = text
	}
	c.answer = strings.TrimSpace(c.answer)
}

func (c *BotContext) AnswerFunc(f func(b *Bot)) {
	c.answerFunc = f
}

func (c *BotContext) Reply(text string, args ...interface{}) {
	c.Text(text, args...)
	c.reply = true
}

func (c *BotContext) Args() []string {
	return c.update.Args()
}

func (c *BotContext) HasArgs() bool {
	return len(c.update.Args()) > 0
}

func (c *BotContext) Arg() string {
	return c.update.Arg()
}

func notFound(c *BotContext) {
	c.Text("Хм, не знаю такую команду")
}

func doRecover() {
	if r := recover(); r != nil {
		fmt.Println("Bot got unexpected error")
		fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
	}
}
