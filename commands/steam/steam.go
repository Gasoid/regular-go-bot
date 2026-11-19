package steam

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/api/playerservice"
	"github.com/gofurry/steam-go/api/steamuser"
	"github.com/gofurry/steam-go/steamid"
	"github.com/gofurry/steam-go/web/community"
	"github.com/gofurry/steam-go/web/market"
)

const (
	appIDCS2  = 730
	appIDDota = 570

	cs2InventoryContextID = "2"

	communityVisibilityPublic = 3
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
		b.WriteString("🔒 profile is private, game stats and inventory are unavailable\n")
		r.SendMessage(b.String())
		return nil
	}

	cs2Hours, dotaHours := fetchPlaytime(ctx, client, steamID)
	fmt.Fprintf(&b, "🎮 CS2: %s · Dota 2: %s\n", formatHours(cs2Hours), formatHours(dotaHours))

	inventoryValue, itemCount, unvaluedCount, invErr := estimateCS2InventoryValue(ctx, client, steamID)
	switch {
	case invErr != nil:
		b.WriteString("💰 CS2 inventory: unavailable\n")
	default:
		fmt.Fprintf(&b, "💰 CS2 inventory: ~$%.2f (%d items", inventoryValue, itemCount)
		if unvaluedCount > 0 {
			fmt.Fprintf(&b, ", %d not priced", unvaluedCount)
		}
		b.WriteString(")\n")
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

func fetchPlaytime(ctx context.Context, client *steam.Client, steamID string) (cs2Minutes, dotaMinutes int) {
	resp, err := client.API.PlayerService.GetOwnedGames(ctx, steamID, &playerservice.GetOwnedGamesOptions{
		AppIDsFilter: []uint32{appIDCS2, appIDDota},
	})
	if err != nil {
		return -1, -1
	}
	cs2Minutes, dotaMinutes = -1, -1
	for _, game := range resp.Response.Games {
		switch game.AppID {
		case appIDCS2:
			cs2Minutes = game.PlaytimeForever
		case appIDDota:
			dotaMinutes = game.PlaytimeForever
		}
	}
	return cs2Minutes, dotaMinutes
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

func estimateCS2InventoryValue(ctx context.Context, client *steam.Client, steamID string) (value float64, itemCount, unvaluedCount int, err error) {
	inventory, err := client.Web.Community.GetInventory(ctx, steamID, appIDCS2, cs2InventoryContextID, &community.GetInventoryOptions{
		Count: 2000,
	})
	if err != nil {
		return 0, 0, 0, err
	}
	if inventory.Success != 1 {
		return 0, 0, 0, errors.New("inventory unavailable: " + inventory.Error)
	}

	descriptionByKey := make(map[string]community.InventoryItem, len(inventory.Descriptions))
	for _, d := range inventory.Descriptions {
		descriptionByKey[d.ClassID+"_"+d.InstanceID] = d
	}

	countByHashName := make(map[string]int)
	for _, asset := range inventory.Assets {
		desc, ok := descriptionByKey[asset.ClassID+"_"+asset.InstanceID]
		if !ok || desc.Marketable != 1 || desc.MarketHashName == "" {
			continue
		}
		amount, convErr := strconv.Atoi(asset.Amount)
		if convErr != nil || amount <= 0 {
			amount = 1
		}
		countByHashName[desc.MarketHashName] += amount
		itemCount += amount
	}

	if len(countByHashName) == 0 {
		return 0, 0, 0, nil
	}

	items := make([]market.PriceOverviewBatchItem, 0, len(countByHashName))
	for name := range countByHashName {
		items = append(items, market.PriceOverviewBatchItem{AppID: appIDCS2, MarketHashName: name})
	}

	results, err := client.Web.Market.GetPriceOverviewBatch(ctx, items, &market.GetPriceOverviewBatchOptions{
		Currency:      1,
		MaxConcurrent: 4,
	})
	if err != nil {
		return 0, itemCount, len(countByHashName), nil
	}

	for _, res := range results {
		if res.Err != nil || !res.Response.Success {
			unvaluedCount += countByHashName[res.Item.MarketHashName]
			continue
		}
		price, priceErr := parsePrice(res.Response.LowestPrice)
		if priceErr != nil {
			unvaluedCount += countByHashName[res.Item.MarketHashName]
			continue
		}
		value += price * float64(countByHashName[res.Item.MarketHashName])
	}

	return value, itemCount, unvaluedCount, nil
}

func parsePrice(raw string) (float64, error) {
	cleaned := strings.NewReplacer("$", "", "€", "", "£", "", ",", "").Replace(strings.TrimSpace(raw))
	if cleaned == "" {
		return 0, errors.New("empty price")
	}
	return strconv.ParseFloat(cleaned, 64)
}

func init() {
	commands.Register(&Command{})
}
