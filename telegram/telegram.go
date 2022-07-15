package telegram

import (
	"log"
	"strings"

	bot "github.com/Gasoid/regular-go-bot/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	timeout = 60
)

type TelegramMessenger struct {
	api      *tgbotapi.BotAPI
	updateCh tgbotapi.UpdatesChannel
}

func (t *TelegramMessenger) Updates() bot.Update {
	return &tgUpdate{update: <-t.updateCh}
}

func (t *TelegramMessenger) BanUser(chatID, userID int64) {
	conf := tgbotapi.BanChatMemberConfig{}
	conf.ChatID = chatID
	conf.UserID = userID
	if _, err := t.api.Send(conf); err != nil {
		log.Println(err.Error())
	}
}

type tgUpdate struct {
	update tgbotapi.Update
	api    *tgbotapi.BotAPI
}

func (tu *tgUpdate) Answer(text string, isReply bool, messageID int) {
	msg := tgbotapi.NewMessage(tu.update.Message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.DisableWebPagePreview = true
	if isReply {
		msg.ReplyToMessageID = messageID
	}
	if _, err := tu.api.Send(msg); err != nil {
		// log.Println(err.Error())
		msg.ParseMode = ""
		log.Println("trying to send it without markdown")
		_, err = tu.api.Send(msg)
		if err != nil {
			log.Println(err.Error())
		}
	}
}

func (tu *tgUpdate) Type() bot.UpdateType {
	if tu.update.Message.IsCommand() {
		return bot.CommandUpdateType
	}
	if tu.update.Message != nil {
		return bot.MessageUpdateType
	}
	return bot.UnknownUpdateType
}

func (tu *tgUpdate) Command() string {
	return tu.update.Message.Command()
}

func (tu *tgUpdate) Args() []string {
	args := tu.update.Message.CommandArguments()
	return strings.Split(args, " ")
}

func (tu *tgUpdate) Arg() string {
	return tu.update.Message.CommandArguments()
}

func (tu *tgUpdate) Message() string {
	return tu.update.Message.Text
}

func (tu *tgUpdate) MessageID() int {
	return tu.update.Message.MessageID
}

func (tu *tgUpdate) User() *bot.User {
	return &bot.User{
		ID:        tu.update.Message.From.ID,
		Username:  tu.update.Message.From.UserName,
		FirstName: tu.update.Message.From.FirstName,
		LastName:  tu.update.Message.From.LastName,
	}
}

func (tu *tgUpdate) NewMembers() []bot.User {
	if tu.update.Message.NewChatMembers == nil {
		return nil
	}
	users := make([]bot.User, len(tu.update.Message.NewChatMembers))
	for i, u := range tu.update.Message.NewChatMembers {
		users[i] = bot.User{
			ID:        u.ID,
			Username:  u.UserName,
			FirstName: u.FirstName,
			LastName:  u.LastName,
		}
	}
	return users
}

func (tu *tgUpdate) ChatID() int64 {
	return tu.update.Message.Chat.ID
}

func (tu *tgUpdate) ChatTitle() string {
	return tu.update.Message.Chat.Title
}

func New(token string) (*TelegramMessenger, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	log.Printf("Authorized on account %s", api.Self.UserName)
	api.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout

	bot := &TelegramMessenger{
		api:      api,
		updateCh: api.GetUpdatesChan(u),
	}
	return bot, nil
}
