package discord

import (
	"context"
	"errors"
	"os"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/channels"
	"github.com/discord-go/discord.go/intents"
)

func init() {
	messenger.RegisterService(new(discordInstance))
}

type chat struct {
	sender   *messenger.Sender
	receiver messenger.Receiver
}

func (c *chat) GetSender() *messenger.Sender {
	return c.sender
}

func (c *chat) GetReceiver() messenger.Receiver {
	return c.receiver
}

type Callback struct {
	sendMessageFunc  func(text string)
	replyMessageFunc func(text string)
}

func (c *Callback) SendMessage(t string) error {
	c.sendMessageFunc(t)
	return nil
}

func (c *Callback) ReplyMessage(t string) error {
	c.replyMessageFunc(t)
	return nil
}

type discordInstance struct{}

//nolint:errcheck
func (t *discordInstance) Run(ctx context.Context, chats chan<- messenger.Chat) error {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return errors.New("DISCORD_TOKEN is required")
	}

	handler := func(ctx *bot.InteractionContext) {

		sender := &messenger.Sender{
			UserID:      ctx.Message.Author.ID.String(),
			Username:    ctx.Message.Author.Username,
			Group:       *ctx.Channel.Name,
			ChatID:      ctx.Message.ChannelID.String(),
			MessageType: messenger.TextMessage,
			Message: messenger.Message{
				IsGroup: ctx.Channel.Type == channels.ChannelTypeGuildText,
				Text:    ctx.Message.Content,
				Command: "",
			},
		}

		switch {
		case ctx.IsCommand():
			sender.Message.Command = ctx.CommandName()
			sender.MessageType = messenger.Command
		case !ctx.IsCommand():
			for _, m := range ctx.Message.Mentions {
				if m.ID == ctx.User().ID {
					sender.MessageType = messenger.Mentioned
					break
				}
			}
		}

		chats <- &chat{
			sender: sender,
			receiver: &messenger.Callback{
				SendMessageFunc: func(text string) {
					ctx.Reply(text)
				},
				ReplyMessageFunc: func(text string) {
					ctx.Reply(text)
				},
			},
		}
	}

	router := bot.NewRouter()
	for _, c := range commands.List() {
		router.Command(c.Name(), c.Help(), handler)
	}

	client := bot.New(token,
		bot.WithIntents(intents.Guilds),
		bot.WithRouter(router),
	)

	if err := client.Run(); err != nil {
		return err
	}

	return nil
}
