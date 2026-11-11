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
