package joke

import (
	"encoding/json"
	"net/http"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
)

const (
	jokeUrl = "https://jokesrv.fermyon.app/oneliner"
)

type Command struct{}

func (c *Command) Name() string {
	return "joke"
}

func (c *Command) Help() string {
	return "the command drops a joke"
}

func (c *Command) Handler(s *messenger.Sender, r messenger.Receiver) error {
	req, err := http.NewRequest("GET", jokeUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	rr := struct {
		Category string `json:"category"`
		Content  string `json:"content"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		return err

	}

	r.SendMessage(rr.Content)
	return nil
}

func init() {
	commands.Register(&Command{})
}
