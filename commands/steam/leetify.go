package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	leetifyApiKeyEnv  = "LEETIFY_API_KEY"
	leetifyApiBaseURL = "https://api-public.cs-prod.leetify.com"

	leetifyPremierRankType = 11
	leetifyPremierHistoryN = 5
	leetifyAvgKillsMatches = 20
	leetifyFormMatches     = 5
)

type leetifyRanks struct {
	Leetify float64 `json:"leetify"`
	Premier int     `json:"premier"`
	Faceit  int     `json:"faceit"`
}

type leetifyProfileResponse struct {
	Winrate      float64      `json:"winrate"`
	TotalMatches int          `json:"total_matches"`
	Ranks        leetifyRanks `json:"ranks"`
}

type leetifyRecentMatch struct {
	FinishedAt string `json:"finished_at"`
	Outcome    string `json:"outcome"`
	Rank       int    `json:"rank"`
	RankType   *int   `json:"rank_type"`
}

type leetifyProfileWithMatchesResponse struct {
	leetifyProfileResponse
	RecentMatches []leetifyRecentMatch `json:"recent_matches"`
}

type leetifyMatchPlayerStats struct {
	Steam64ID  string `json:"steam64_id"`
	TotalKills int    `json:"total_kills"`
}

type leetifyMatch struct {
	FinishedAt string                    `json:"finished_at"`
	Stats      []leetifyMatchPlayerStats `json:"stats"`
}

type leetifyStats struct {
	profile *leetifyProfileWithMatchesResponse
	// premierHistory holds the most recent non-zero premier ranks, newest first,
	// used as a fallback when the player has no rank for the current season.
	premierHistory []leetifyRecentMatch
	avgKills       float64
	avgKillsOK     bool
	// form holds outcomes of the last leetifyFormMatches matches, oldest first.
	form []string
}

func fetchLeetifyStats(ctx context.Context, steamID string) (*leetifyStats, error) {
	profile, err := fetchLeetifyProfile(ctx, steamID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}

	stats := &leetifyStats{profile: profile}
	stats.premierHistory = premierHistoryFromMatches(profile.RecentMatches)
	stats.form = formFromMatches(profile.RecentMatches, leetifyFormMatches)

	matches, err := fetchLeetifyMatches(ctx, steamID)
	if err == nil {
		if avg, ok := avgKillsFromMatches(matches, steamID, leetifyAvgKillsMatches); ok {
			stats.avgKills = avg
			stats.avgKillsOK = true
		}
	}

	return stats, nil
}

func premierHistoryFromMatches(matches []leetifyRecentMatch) []leetifyRecentMatch {
	history := make([]leetifyRecentMatch, 0, leetifyPremierHistoryN)
	for _, m := range matches {
		if m.RankType == nil || *m.RankType != leetifyPremierRankType || m.Rank <= 0 {
			continue
		}
		history = append(history, m)
		if len(history) >= leetifyPremierHistoryN {
			break
		}
	}
	return history
}

// formFromMatches returns up to n outcomes ("win"/"loss"/"draw") of the most
// recent matches, oldest first, so callers can render them left-to-right in
// chronological order.
func formFromMatches(matches []leetifyRecentMatch, n int) []string {
	if len(matches) > n {
		matches = matches[:n]
	}
	form := make([]string, 0, len(matches))
	for i := len(matches) - 1; i >= 0; i-- {
		form = append(form, matches[i].Outcome)
	}
	return form
}

func avgKillsFromMatches(matches []leetifyMatch, steamID string, limit int) (float64, bool) {
	var total, count int
	for _, m := range matches {
		if count >= limit {
			break
		}
		for _, s := range m.Stats {
			if s.Steam64ID == steamID {
				total += s.TotalKills
				count++
				break
			}
		}
	}
	if count == 0 {
		return 0, false
	}
	return float64(total) / float64(count), true
}

//nolint:errcheck
func fetchLeetifyProfile(ctx context.Context, steamID string) (*leetifyProfileWithMatchesResponse, error) {
	url := fmt.Sprintf("%s/v3/profile?steam64_id=%s", leetifyApiBaseURL, steamID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	setLeetifyAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetify api returned status %d", resp.StatusCode)
	}

	var profile leetifyProfileWithMatchesResponse
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

//nolint:errcheck
func fetchLeetifyMatches(ctx context.Context, steamID string) ([]leetifyMatch, error) {
	url := fmt.Sprintf("%s/v3/profile/matches?steam64_id=%s", leetifyApiBaseURL, steamID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	setLeetifyAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetify api returned status %d", resp.StatusCode)
	}

	var matches []leetifyMatch
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {
		return nil, err
	}
	return matches, nil
}

func formatForm(form []string) string {
	var b strings.Builder
	for _, outcome := range form {
		switch outcome {
		case "win":
			b.WriteString("🟢")
		case "loss":
			b.WriteString("🔴")
		default:
			b.WriteString("⚪")
		}
	}
	return b.String()
}

func setLeetifyAuth(req *http.Request) {
	if leetifyKey := os.Getenv(leetifyApiKeyEnv); leetifyKey != "" {
		req.Header.Set("Authorization", "Bearer "+leetifyKey)
	}
}
