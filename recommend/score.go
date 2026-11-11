package recommend

import (
	"gamematch/store"
	"sort"
	"strings"
)

type Preference struct {
	Genre string `json:"genre"`
	Mode  string `json:"mode"`
	Style string `json:"style"`
}

func contains(xs []string, opts ...string) bool {
	for _, x := range xs {
		for _, y := range opts {
			if strings.EqualFold(x, y) {
				return true
			}
		}
	}
	return false
}
func Valid(p Preference) bool {
	return contains([]string{"rpg", "action", "strategy", "simulation", "adventure"}, p.Genre) && contains([]string{"solo", "coop", "pvp"}, p.Mode) && contains([]string{"story", "competitive", "relax", "challenge"}, p.Style)
}
func Rank(games []store.Game, p Preference) []store.Game {
	out := []store.Game{}
	for _, g := range games {
		if !g.Published {
			continue
		}
		g.Score = 0
		g.Reasons = []string{}
		if contains(g.Genres, p.Genre) || (p.Genre == "rpg" && contains(g.Genres, "role-playing-games-rpg")) {
			g.Score += 45
			g.Reasons = append(g.Reasons, "แนวเกมตรงกับความชอบ +45")
		}
		modes := map[string][]string{"solo": {"singleplayer", "single-player"}, "coop": {"co-op", "cooperative", "online-co-op", "local-co-op"}, "pvp": {"pvp", "online-pvp", "local-pvp"}}
		if contains(g.Tags, modes[p.Mode]...) {
			g.Score += 25
			g.Reasons = append(g.Reasons, "รูปแบบการเล่นตรง +25")
		}
		styles := map[string][]string{"story": {"story-rich", "story-driven"}, "competitive": {"competitive", "pvp"}, "relax": {"relaxing", "casual"}, "challenge": {"difficult", "challenging", "souls-like"}}
		if contains(g.Tags, styles[p.Style]...) {
			g.Score += 30
			g.Reasons = append(g.Reasons, "สไตล์การเล่นตรง +30")
		}
		if g.Score > 0 {
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			if out[i].Title == out[j].Title {
				return out[i].ID < out[j].ID
			}
			return out[i].Title < out[j].Title
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}
