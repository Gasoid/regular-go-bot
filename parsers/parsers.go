package parsers

type Parser interface {
	Handler(string, Callback) error
}

var (
	voiceParsers = []Parser{}
)

func RegisterVoiceParser(parser Parser) {
	voiceParsers = append(voiceParsers, parser)
}

func ListVoiceParsers() []Parser {
	return voiceParsers
}

type Callback struct {
	SendMessage  func(text string)
	SendVideo    func(filePath string)
	ReplyMessage func(text string)
}
