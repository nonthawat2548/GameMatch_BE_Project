package tests

import (
	"gamematch/recommend"
	"gamematch/store"
	"testing"
)

func TestMatchAndExcludeUnpublished(t *testing.T) {
	p := recommend.Preference{Genre: "rpg", Mode: "solo", Style: "story"}
	games := []store.Game{{ID: 1, Title: "A", Published: true, Genres: []string{"rpg"}, Tags: []string{"singleplayer", "story-rich"}}, {ID: 2, Title: "hidden", Published: false, Genres: []string{"rpg"}}}
	got := recommend.Rank(games, p)
	if len(got) != 1 || got[0].Score != 100 {
		t.Fatalf("unexpected %#v", got)
	}
}
func TestMultiplayerIsNotCoop(t *testing.T) {
	g := recommend.Rank([]store.Game{{Published: true, Genres: []string{"action"}, Tags: []string{"multiplayer"}}}, recommend.Preference{Genre: "action", Mode: "coop", Style: "relax"})
	if len(g) != 1 || g[0].Score != 45 {
		t.Fatal(g)
	}
}
func TestEmptyAndStableTie(t *testing.T) {
	if len(recommend.Rank(nil, recommend.Preference{})) != 0 {
		t.Fatal("empty")
	}
	g := recommend.Rank([]store.Game{{Title: "Z", Published: true, Genres: []string{"rpg"}}, {Title: "A", Published: true, Genres: []string{"rpg"}}}, recommend.Preference{Genre: "rpg"})
	if g[0].Title != "A" {
		t.Fatal(g)
	}
}

func TestExpandedGenresAndBreakdown(t *testing.T) {
	p := recommend.Preference{Genre: "horror", Mode: "coop", Style: "challenge"}
	if !recommend.Valid(p) {
		t.Fatal("horror should be valid")
	}
	games := []store.Game{{ID: 1, Published: true, Genres: []string{"action"}, Tags: []string{"horror", "co-op"}}}
	got := recommend.Rank(games, p)
	if len(got) != 1 || got[0].Score != 70 || len(got[0].MatchBreakdown) != 3 || got[0].MatchBreakdown[2].Points != 0 {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestSimilarExcludesReferencesAndHiddenGames(t *testing.T) {
	ref := store.Game{ID: 1, Title: "Reference", Published: true, Genres: []string{"action"}, Tags: []string{"fps"}}
	games := []store.Game{
		ref,
		{ID: 2, Title: "Similar", Published: true, Genres: []string{"action"}, Tags: []string{"fps"}},
		{ID: 3, Title: "Hidden", Published: false, Genres: []string{"action"}},
		{ID: 4, Title: "Unrelated", Published: true, Genres: []string{"sports"}},
	}
	got := recommend.Similar(games, []store.Game{ref})
	if len(got) != 1 || got[0].ID != 2 || got[0].SimilarityScore != 4 || got[0].Score != 0 {
		t.Fatalf("unexpected result: %#v", got)
	}
	if len(recommend.Similar(games, nil)) != 0 {
		t.Fatal("no references should yield no recommendations")
	}
}
