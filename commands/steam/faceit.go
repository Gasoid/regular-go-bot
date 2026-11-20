package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	faceitApiKeyEnv  = "FACEIT_API_KEY"
	faceitApiBaseURL = "https://open.faceit.com/data/v4"
	faceitCS2Game    = "cs2"
)

type faceitGameStats struct {
	SkillLevel int `json:"skill_level"`
	FaceitElo  int `json:"faceit_elo"`
}

type faceitPlayerResponse struct {
	Games map[string]faceitGameStats `json:"games"`
}

//nolint:errcheck
func fetchFaceitStats(ctx context.Context, apiKey, steamID string) (*faceitGameStats, error) {
	url := fmt.Sprintf("%s/players?game=%s&game_player_id=%s", faceitApiBaseURL, faceitCS2Game, steamID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("faceit api returned status %d", resp.StatusCode)
	}

	var player faceitPlayerResponse
	if err := json.NewDecoder(resp.Body).Decode(&player); err != nil {
		return nil, err
	}

	cs2Stats, ok := player.Games[faceitCS2Game]
	if !ok {
		return nil, nil
	}
	return &cs2Stats, nil
}
