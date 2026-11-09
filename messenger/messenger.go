package messenger

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
)

type Message struct {
	IsGroupEvent bool
	MessageType  EventType
	Text         string
}

type Sender struct {
	UserID   string
	Username string
	Message  Message
	Group    string
}

type Receiver interface {
	SendMessage(text string) error
}

type Chat interface {
	GetSender() *Sender
	GetReceiver() Receiver
}
