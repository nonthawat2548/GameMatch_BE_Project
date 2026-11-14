package store

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Keep the first occurrence and order so the public JSON array remains stable.
func cleanGenres(genres []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, slug := range genres {
		slug = strings.TrimSpace(slug)
		if slug == "" || seen[slug] {
			continue
		}
		seen[slug] = true
		out = append(out, slug)
	}
	return out
}

// All writes update the many-to-many relation inside the same transaction.
// The existing array stays in sync for backward compatibility and safe rollback.
func replaceGameGenres(ctx context.Context, tx pgx.Tx, gameID int64, genres []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM gm_game_genres WHERE game_id=$1`, gameID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO gm_genres(slug) SELECT DISTINCT unnest($1::text[]) ORDER BY 1 ON CONFLICT (slug) DO NOTHING`, genres); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO gm_game_genres(game_id,genre_id,position) SELECT $1,d.id,x.pos::integer FROM unnest($2::text[]) WITH ORDINALITY AS x(slug,pos) JOIN gm_genres d ON d.slug=x.slug`, gameID, genres)
	return err
}
