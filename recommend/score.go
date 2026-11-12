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
// Normalize keeps the accepted input and saved values in the same format.
func Normalize(p Preference) Preference {
	return Preference{
		Genre: strings.ToLower(strings.TrimSpace(p.Genre)),
		Mode: strings.ToLower(strings.TrimSpace(p.Mode)),
		Style: strings.ToLower(strings.TrimSpace(p.Style)),
	}
}

func Valid(p Preference) bool {
	p = Normalize(p)
	return contains([]string{"rpg", "action", "strategy", "simulation", "adventure", "indie", "casual", "sports", "racing", "massively-multiplayer", "early-access", "free-to-play", "fps", "survival", "horror", "puzzle", "moba"}, p.Genre) && contains([]string{"solo", "coop", "pvp"}, p.Mode) && contains([]string{"story", "competitive", "relax", "challenge"}, p.Style)
}
func Rank(games []store.Game, p Preference) []store.Game {
	p = Normalize(p)
	out := []store.Game{}
	for _, g := range games {
		if !g.Published {
			continue
		}
		g.Score = 0
		g.Reasons = []string{}
		g.MatchBreakdown = []store.MatchPart{
			{Label: "แนวเกม", Max: 45},
			{Label: "รูปแบบการเล่น", Max: 25},
			{Label: "สไตล์การเล่น", Max: 30},
		}
		tagGenre := contains([]string{"fps", "survival", "horror", "puzzle", "moba"}, p.Genre)
		if contains(g.Genres, p.Genre) || (p.Genre == "rpg" && contains(g.Genres, "role-playing-games-rpg")) || (tagGenre && contains(g.Tags, p.Genre)) {
			g.MatchBreakdown[0].Points = 45
			g.Score += 45
			g.Reasons = append(g.Reasons, "แนวเกมตรงกับความชอบ +45")
		}
		modes := map[string][]string{"solo": {"singleplayer", "single-player"}, "coop": {"co-op", "cooperative", "online-co-op", "local-co-op", "co-op-campaign"}, "pvp": {"pvp", "online-pvp", "local-pvp"}}
		if contains(g.Tags, modes[p.Mode]...) {
			g.MatchBreakdown[1].Points = 25
			g.Score += 25
			g.Reasons = append(g.Reasons, "รูปแบบการเล่นตรง +25")
		}
		styles := map[string][]string{"story": {"story-rich", "story-driven"}, "competitive": {"competitive", "pvp"}, "relax": {"relaxing", "casual", "cozy"}, "challenge": {"difficult", "challenging", "souls-like"}}
		if contains(g.Tags, styles[p.Style]...) {
			g.MatchBreakdown[2].Points = 30
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
	if len(out) > 10 {
		out = out[:10]
	}
	return out
}

// Similar recommends published games based on shared genres and gameplay tags.
// Each candidate uses its best matching reference, not the number of favorites.
func Similar(games []store.Game, references []store.Game) []store.Game {
	out := []store.Game{}
	excluded := map[int64]bool{}
	for _, ref := range references {
		excluded[ref.ID] = true
	}
	for _, g := range games {
		if !g.Published || excluded[g.ID] {
			continue
		}
		best := 0
		var reasons []string
		for _, ref := range references {
			genres := common(g.Genres, ref.Genres)
			tags := common(g.Tags, ref.Tags)
			points := 3*len(genres) + len(tags)
			if points > best {
				best = points
				reasons = []string{"คล้ายกับ " + ref.Title}
				if len(genres) > 0 {
					reasons = append(reasons, "แนวเกมร่วม: " + strings.Join(genres, ", "))
				}
				if len(tags) > 0 {
					if len(tags) > 6 {
						tags = tags[:6]
					}
					reasons = append(reasons, "แท็กร่วม: " + strings.Join(tags, ", "))
				}
			}
		}
		if best > 0 {
			g.SimilarityScore = best
			g.Score = 0
			g.MatchBreakdown = nil
			g.Reasons = reasons
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SimilarityScore == out[j].SimilarityScore {
			if out[i].Title == out[j].Title {
				return out[i].ID < out[j].ID
			}
			return out[i].Title < out[j].Title
		}
		return out[i].SimilarityScore > out[j].SimilarityScore
	})
	if len(out) > 4 {
		out = out[:4]
	}
	return out
}

func common(a, b []string) []string {
	out := []string{}
	for _, x := range a {
		// Free-to-play and Early Access are business/release labels, not gameplay.
		if contains([]string{"free-to-play", "early-access", "indie"}, x) {
			continue
		}
		if contains(b, x) && !contains(out, x) {
			out = append(out, x)
		}
	}
	return out
}
