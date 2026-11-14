package discord

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/discord-go/discord.go/bot"
	"github.com/discord-go/discord.go/intents"
	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
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

type discordInstance struct{}

//nolint:errcheck
func (t *discordInstance) Run(ctx context.Context, chats chan<- messenger.Chat) error {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		slog.Info("No DISCORD_TOKEN provided, so discord has not started")
		return nil
	}

	handler := func(ctx *bot.InteractionContext) {

		if !ctx.IsCommand() {
			return
		}

		args := []string{}
		for _, o := range ctx.Options() {
			args = append(args, ctx.OptionString(o.Name))
		}

		user := ctx.User()

		sender := &messenger.Sender{
			UserID:      user.ID.String(),
			Username:    user.Username,
			Group:       *ctx.Channel.Name,
			ChatID:      ctx.ChannelID().String(),
			MessageType: messenger.Command,
			Message: messenger.Message{
				IsGroup: ctx.InGuild(),
				Command: ctx.CommandName(),
				Arg:     strings.Join(args, " "),
			},
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

	messageHandler := func(ctx *bot.MessageContext) {
		groupName := ""
		if ctx.Thread != nil {
			groupName = *ctx.Thread.Name
		}

		sender := &messenger.Sender{
			UserID:      ctx.Author.ID.String(),
			Username:    ctx.Author.Username,
			Group:       groupName,
			ChatID:      ctx.Message.ChannelID.String(),
			MessageType: messenger.TextMessage,
			Message: messenger.Message{
				IsGroup: !ctx.GuildID.IsZero(),
				Text:    ctx.Content,
			},
		}

		switch {
		case len(ctx.Attachments) > 0:
			for _, a := range ctx.Attachments {
				if !strings.HasPrefix(a.ContentType, "audio/") || a.DurationSecs == 0 || a.Waveform == "" {
					continue
				}

				sender.Message.VoiceUrl = a.URL
				sender.MessageType = messenger.VoiceMessage
				break
			}
		case len(ctx.Mentions) > 0:
			for _, m := range ctx.Mentions {
				if m.ID == ctx.Bot.User().ID {
					sender.MessageType = messenger.Mentioned
					break
				}
			}
			fallthrough
		case ctx.ReferencedMessage != nil && ctx.ReferencedMessage.Type == messages.MessageTypeReply:
			sender.MessageType = messenger.ReplyMessage
			sender.Message.ReplyToMessage = ctx.ReferencedMessage.Content
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
				SendTypingFunc: func() {
					if err := ctx.TriggerTyping(); err != nil {
						slog.Error("counldn't send typing", "err", err)
					}
				},
			},
		}
	}

	router := bot.NewRouter()

	for _, c := range commands.List() {
		options := []interactions.ApplicationCommandOption{}
		for _, a := range c.Args() {
			options = append(options, interactions.ApplicationCommandOption{
				Type:        interactions.ApplicationCommandOptionTypeString,
				Name:        a,
				Description: "optional",
			})
		}

		router.Command(
			c.Name(),
			c.Help(),
			handler,
			options...)
	}

	client := bot.New(token,
		bot.WithIntents(intents.Guilds|intents.GuildMessages|intents.MessageContent),
		bot.WithRouter(router),
	)

	client.OnMessageCreate(messageHandler)

	slog.Info("discord service is running")

	if err := client.RunContext(ctx); err != nil {
		return err
	}

	return nil
}
