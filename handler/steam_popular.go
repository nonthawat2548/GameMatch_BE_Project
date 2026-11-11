package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"gamematch/store"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type steamCandidate struct {
	ID     int64
	Source string
}

var steamSearchIDs = regexp.MustCompile(`data-ds-appid="(\d+)"`)

func popularCandidates(ctx context.Context, client *http.Client) ([]steamCandidate, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.steampowered.com/ISteamChartsService/GetMostPlayedGames/v1/", nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	var chart struct {
		Response struct {
			Ranks []struct {
				Rank int   `json:"rank"`
				ID   int64 `json:"appid"`
			} `json:"ranks"`
		} `json:"response"`
	}
	if err != nil || res.StatusCode != 200 || len(data) > 1<<20 || json.Unmarshal(data, &chart) != nil || len(chart.Response.Ranks) == 0 || len(chart.Response.Ranks) > 1000 {
		return nil, fmt.Errorf("Steam Most Played unavailable")
	}
	sort.Slice(chart.Response.Ranks, func(i, j int) bool { return chart.Response.Ranks[i].Rank < chart.Response.Ranks[j].Rank })
	items := []steamCandidate{}
	seen := map[int64]bool{}
	add := func(id int64, source string) {
		if id > 0 && !seen[id] {
			seen[id] = true
			items = append(items, steamCandidate{id, source})
		}
	}
	for i, row := range chart.Response.Ranks {
		if i >= 100 {
			break
		}
		add(row.ID, "Steam Most Played")
	}
	// Top Sellers supplies eligible replacements for software/private/unavailable chart entries.
	for _, start := range []int{0, 50, 100, 150} {
		body, err := steamFetch(ctx, client, "/search/results/?start="+strconv.Itoa(start)+"&count=50&filter=topsellers&category1=998&infinite=1&cc=th&l=english")
		var search struct {
			Success int    `json:"success"`
			HTML    string `json:"results_html"`
		}
		if err != nil || json.Unmarshal(body, &search) != nil || search.Success != 1 {
			return nil, fmt.Errorf("Steam Top Sellers unavailable")
		}
		matches := steamSearchIDs.FindAllStringSubmatch(search.HTML, 100)
		if len(matches) == 0 {
			return nil, fmt.Errorf("Steam Top Sellers format changed")
		}
		for _, match := range matches {
			id, _ := strconv.ParseInt(match[1], 10, 64)
			add(id, "Steam Top Sellers (TH)")
		}
	}
	return items, nil
}

// ImportSteamPopular is an explicit maintenance command, not a long-running HTTP request.
// Fetch all eligible metadata first, then commit the selected catalog in one transaction.
func ImportSteamPopular(ctx context.Context, s *store.Store, client *http.Client, limit int, progress func(int, int, string)) (int, error) {
	if limit < 1 || limit > 100 {
		return 0, fmt.Errorf("popular catalog limit must be 1..100")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	candidates, err := popularCandidates(ctx, client)
	if err != nil {
		return 0, err
	}
	games := []store.Game{}
	pace := time.NewTicker(400 * time.Millisecond)
	defer pace.Stop()
	for index, candidate := range candidates {
		if len(games) >= limit {
			break
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-pace.C:
		}
		id := strconv.FormatInt(candidate.ID, 10)
		body, err := steamFetch(ctx, client, "/api/appdetails?appids="+id+"&cc=th&l=english")
		var response map[string]struct {
			Success bool        `json:"success"`
			Data    steamDetail `json:"data"`
		}
		if err != nil || json.Unmarshal(body, &response) != nil {
			return 0, fmt.Errorf("Steam details unavailable for app %s", id)
		}
		entry, ok := response[id]
		if !ok || !entry.Success || entry.Data.ID != candidate.ID || entry.Data.Type != "game" || entry.Data.Release.Soon {
			continue
		}
		page, err := steamFetch(ctx, client, "/app/"+id+"/?l=english&cc=th")
		if err != nil {
			return 0, fmt.Errorf("Steam tags unavailable for app %s", id)
		}
		game, ok := steamGame(entry.Data, page)
		if !ok {
			continue
		}
		position := len(games) + 1
		game.PopularityOrder = &position
		game.PopularitySource = candidate.Source
		games = append(games, game)
		if progress != nil {
			progress(len(games), index+1, game.Title)
		}
	}
	if len(games) < limit {
		return 0, fmt.Errorf("only %d/%d eligible games; previous catalog unchanged", len(games), limit)
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	// Reset ranking metadata only after all requested records have been verified.
	if _, err = tx.Exec(ctx, `UPDATE gm_games SET popularity_order=NULL,popularity_source=NULL,popularity_updated_at=NULL WHERE popularity_order IS NOT NULL`); err != nil {
		return 0, err
	}
	for _, g := range games {
		_, err = tx.Exec(ctx, `INSERT INTO gm_games(steam_id,title,description,image_url,genres,tags,platforms,release_date,synced_at,popularity_order,popularity_source,popularity_updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8::date,now(),$9,$10,now()) ON CONFLICT(steam_id) DO UPDATE SET title=EXCLUDED.title,image_url=EXCLUDED.image_url,genres=EXCLUDED.genres,tags=EXCLUDED.tags,platforms=EXCLUDED.platforms,release_date=EXCLUDED.release_date,synced_at=EXCLUDED.synced_at,popularity_order=EXCLUDED.popularity_order,popularity_source=EXCLUDED.popularity_source,popularity_updated_at=EXCLUDED.popularity_updated_at`, g.SteamID, g.Title, g.Description, g.ImageURL, g.Genres, g.Tags, g.Platforms, g.ReleaseDate, g.PopularityOrder, g.PopularitySource)
		if err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(games), nil
}
