package parsers

import (
	"time"

	"github.com/Gasoid/regular-go-bot/metrics"
)

type Parser interface {
	Handler(string, Callback) error
	Name() string
}

var (
	voiceParsers    = []Parser{}
	locationParsers = []Parser{}
	textParsers     = []Parser{}
)

type Wrapper struct {
	parser Parser
}

func (w *Wrapper) Handler(s string, c Callback) error {
	start := time.Now()
	err := w.parser.Handler(s, c)
	metrics.ParserInc(w.parser.Name(), err)
	metrics.ParserDuration(w.parser.Name(), time.Since(start))
	return err
}

func (w *Wrapper) Name() string {
	return w.parser.Name()
}

func RegisterVoiceParser(parser Parser) {
	voiceParsers = append(voiceParsers, &Wrapper{parser})
}

func ListVoiceParsers() []Parser {
	return voiceParsers
}

func RegisterLocationParser(parser Parser) {
	locationParsers = append(locationParsers, &Wrapper{parser})
}

func ListLocationParsers() []Parser {
	return locationParsers
}

func RegisterTextParser(parser Parser) {
	textParsers = append(textParsers, &Wrapper{parser})
}

func ListTextParsers() []Parser {
	return textParsers
}

type Callback struct {
	SendMessage  func(text string)
	SendVideo    func(filePath string)
	SendPhoto    func(filePath, caption string)
	ReplyMessage func(text string)
}
