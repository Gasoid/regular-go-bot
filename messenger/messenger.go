package messenger

import "context"

type HandlerFunc func(sender *Sender, receiver Receiver) error

func (f HandlerFunc) Process(s *Sender, r Receiver) error {
	return f(s, r)
}

type Handler interface {
	Process(s *Sender, r Receiver) error
}

type EventType uint8

const (
	TextMessage EventType = iota
	NewMember
	VoiceMessage
	VideoMessage
	Command
	Location
	ReplyMessage
	Mentioned
	Photo
)

type Message struct {
	IsGroup        bool
	Text           string
	Command        string
	VoiceFilePath  string
	Coordinates    [2]float64
	Arg            string
	PhotoPaths     []string
	ReplyToMessage string
}

type Sender struct {
	UserID      string
	Username    string
	Message     Message
	Group       string
	ChatID      string
	MessageType EventType
}

type Receiver interface {
	SendMessage(text string) error
	ReplyMessage(text string) error
}

type Chat interface {
	GetSender() *Sender
	GetReceiver() Receiver
}

type Service interface {
	Run(context.Context, chan<- Chat) error
}

var (
	services []Service
)

func RegisterService(s Service) {
	services = append(services, s)
}

func ListServices() []Service {
	return services
}
