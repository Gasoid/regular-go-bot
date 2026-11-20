package steam

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/api/playerservice"
	"github.com/gofurry/steam-go/api/steamuser"
	"github.com/gofurry/steam-go/steamid"
)

const (
	communityVisibilityPublic = 3

	topGamesCount = 3
)

type Command struct{}

func (c *Command) Name() string {
	return "steam"
}

func (c *Command) Help() string {
	return "the command returns steam account statistics, e.g.: /steam https://steamcommunity.com/id/example"
}

func (c *Command) Args() []string {
	return []string{"steam-url"}
}

//nolint:errcheck
func (c *Command) Handler(s *messenger.Sender, r messenger.Receiver) error {
	steamApiKey := os.Getenv("STEAM_API_KEY")
	if steamApiKey == "" {
		return errors.New("STEAM_API_KEY is unset, you need to set up STEAM_API_KEY var")
	}

	if s.Message.Arg == "" {
		r.ReplyMessage("profile id is not provided")
		return nil
	}

	client, err := steam.NewClient(
		steam.WithAPIKey(steamApiKey),
		steam.WithTimeout(15*time.Second),
		steam.WithRetry(2),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := context.Background()

	steamID, err := resolveSteamID(ctx, client, s.Message.Arg)
	if err != nil {
		r.ReplyMessage("couldn't find a steam profile for this input")
		return nil
	}

	summary, bans, err := fetchCoreInfo(ctx, client, steamID)
	if err != nil {
		return err
	}
	if summary == nil {
		r.ReplyMessage("steam profile not found")
		return nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "👤 %s\n", summary.PersonaName)
	if summary.LocCountryCode != "" {
		fmt.Fprintf(&b, "%s %s\n", countryFlag(summary.LocCountryCode), summary.LocCountryCode)
	}
	if summary.TimeCreated > 0 {
		createdAt := time.Unix(summary.TimeCreated, 0)
		age := int(time.Since(createdAt).Hours() / 24 / 365)
		fmt.Fprintf(&b, "📅 created: %s (%d years)\n", createdAt.Format("2006-01-02"), age)
	}
	if bans != nil {
		fmt.Fprintf(&b, "🚫 bans: VAC %s, game bans: %d, economy: %s\n",
			yesNo(bans.VACBanned), bans.NumberOfGameBans, bans.EconomyBan)
	}

	if summary.CommunityVisibilityState != communityVisibilityPublic {
		b.WriteString("🔒 profile is private, game stats are unavailable\n")
		r.SendMessage(b.String())
		return nil
	}

	topGames, gamesErr := fetchTopGames(ctx, client, steamID)
	switch {
	case gamesErr != nil:
		b.WriteString("🎮 top games: unavailable\n")
	case len(topGames) == 0:
		b.WriteString("🎮 top games: no games in library\n")
	default:
		b.WriteString("🎮 top games:\n")
		for _, g := range topGames {
			fmt.Fprintf(&b, "  • %s: %s\n", g.Name, formatHours(g.PlaytimeForever))
		}
	}

	if faceitKey := os.Getenv(faceitApiKeyEnv); faceitKey != "" {
		if faceit, err := fetchFaceitStats(ctx, faceitKey, steamID); err == nil && faceit != nil {
			fmt.Fprintf(&b, "🟠 Faceit: level %d, elo %d\n", faceit.SkillLevel, faceit.FaceitElo)
		}
	}

	if leetify, err := fetchLeetifyStats(ctx, steamID); err == nil && leetify != nil {
		fmt.Fprintf(&b, "📊 Leetify: rating %.2f, winrate %.1f%% (%d matches)\n",
			leetify.profile.Ranks.Leetify, leetify.profile.Winrate*100, leetify.profile.TotalMatches)

		switch {
		case leetify.profile.Ranks.Premier > 0:
			fmt.Fprintf(&b, "  premier: %d\n", leetify.profile.Ranks.Premier)
		case len(leetify.premierHistory) > 0:
			b.WriteString("  premier (past seasons):")
			for _, m := range leetify.premierHistory {
				fmt.Fprintf(&b, " %d", m.Rank)
			}
			b.WriteString("\n")
		}

		if leetify.avgKillsOK {
			fmt.Fprintf(&b, "  avg kills (last %d matches): %.1f\n", leetifyAvgKillsMatches, leetify.avgKills)
		}

		if len(leetify.form) > 0 {
			fmt.Fprintf(&b, "  form: %s\n", formatForm(leetify.form))
		}
	}

	r.SendMessage(b.String())
	return nil
}

func resolveSteamID(ctx context.Context, client *steam.Client, arg string) (string, error) {
	arg = strings.TrimSpace(arg)

	if id, err := steamid.ParseCommunityURL(arg); err == nil {
		return id.String(), nil
	} else if errors.Is(err, steamid.ErrVanityReference) {
		vanity := strings.Trim(arg, "/")
		if _, after, ok := strings.CutLast(vanity, "/"); ok {
			vanity = after
		}
		return resolveVanity(ctx, client, vanity)
	}

	if id, err := steamid.Parse(arg); err == nil {
		return id.String(), nil
	}

	vanity := strings.Trim(arg, "/")
	if _, after, ok := strings.CutLast(vanity, "/"); ok {
		vanity = after
	}
	return resolveVanity(ctx, client, vanity)
}

func resolveVanity(ctx context.Context, client *steam.Client, vanity string) (string, error) {
	if vanity == "" {
		return "", errors.New("empty vanity name")
	}
	resp, err := client.API.SteamUser.ResolveVanityURL(ctx, vanity, nil)
	if err != nil {
		return "", err
	}
	if resp.Response.Success != 1 || resp.Response.SteamID == "" {
		return "", errors.New("vanity name not resolved")
	}
	return resp.Response.SteamID, nil
}

func fetchCoreInfo(ctx context.Context, client *steam.Client, steamID string) (*steamuser.Player, *steamuser.PlayerBan, error) {
	summaries, err := client.API.SteamUser.GetPlayerSummaries(ctx, []string{steamID})
	if err != nil {
		return nil, nil, err
	}
	if len(summaries.Response.Players) == 0 {
		return nil, nil, nil
	}
	player := summaries.Response.Players[0]

	bansResp, err := client.API.SteamUser.GetPlayerBans(ctx, []string{steamID})
	if err != nil {
		return &player, nil, nil
	}
	var bans *steamuser.PlayerBan
	if len(bansResp.Players) > 0 {
		bans = &bansResp.Players[0]
	}

	return &player, bans, nil
}

func fetchTopGames(ctx context.Context, client *steam.Client, steamID string) ([]playerservice.OwnedGame, error) {
	resp, err := client.API.PlayerService.GetOwnedGames(ctx, steamID, nil)
	if err != nil {
		return nil, err
	}

	games := resp.Response.Games
	sort.Slice(games, func(i, j int) bool {
		return games[i].PlaytimeForever > games[j].PlaytimeForever
	})

	if len(games) > topGamesCount {
		games = games[:topGamesCount]
	}
	return games, nil
}

func formatHours(minutes int) string {
	if minutes < 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f h", float64(minutes)/60)
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func countryFlag(countryCode string) string {
	if len(countryCode) != 2 {
		return "🏳️"
	}
	code := strings.ToUpper(countryCode)
	first := code[0]
	second := code[1]
	if first < 'A' || first > 'Z' || second < 'A' || second > 'Z' {
		return "🏳️"
	}
	return string(rune(0x1F1E6+int(first-'A'))) + string(rune(0x1F1E6+int(second-'A')))
}

func init() {
	commands.Register(&Command{})
}
