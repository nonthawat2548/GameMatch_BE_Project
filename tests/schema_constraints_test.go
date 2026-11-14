package tests

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const identityMigrationPath = "../supabase/migrations/20261113151856_enforce_member_identity_and_review_constraints.sql"

func requireSQLState(t *testing.T, err error, code string) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != code {
		t.Fatalf("expected SQLSTATE %s, got %v", code, err)
	}
}

func TestSchemaIdentityAndTextConstraints(t *testing.T) {
	a, _ := databaseAPI(t)
	ctx := context.Background()
	var memberID, gameID int64
	if err := a.S.DB.QueryRow(ctx, `INSERT INTO gm_members(email,username,password_hash,display_name) VALUES('owner@example.test','Owner','test-only-hash','Owner') RETURNING id`).Scan(&memberID); err != nil {
		t.Fatal(err)
	}
	if err := a.S.DB.QueryRow(ctx, `INSERT INTO gm_games(title) VALUES('Constraint fixture') RETURNING id`).Scan(&gameID); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, username, displayName, code string
	}{
		{"duplicate username", "Owner", "Other", "23505"},
		{"foreign email username", "victim@example.test", "Other", "23514"},
		{"blank name", "Blank", "   ", "23514"},
		{"long name", "Long", strings.Repeat("ก", 51), "23514"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := a.S.DB.Exec(ctx, `INSERT INTO gm_members(email,username,password_hash,display_name) VALUES('other@example.test',$1,'test-only-hash',$2)`, tc.username, tc.displayName)
			requireSQLState(t, err, tc.code)
		})
	}
	for _, tc := range []struct{ name, comment string }{
		{"short comment", "123456789"},
		{"blank comment", strings.Repeat(" ", 20)},
		{"long UTF-8 comment", strings.Repeat("ก", 667)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := a.S.DB.Exec(ctx, `INSERT INTO gm_reviews(member_id,game_id,rating,comment) VALUES($1,$2,5,$3)`, memberID, gameID, tc.comment)
			requireSQLState(t, err, "23514")
		})
	}
	// Valid limits use characters for display names and bytes for review text.
	if _, err := a.S.DB.Exec(ctx, `UPDATE gm_members SET display_name=$1 WHERE id=$2`, strings.Repeat("ก", 50), memberID); err != nil {
		t.Fatal("valid 50-character name rejected: ", err)
	}
	if _, err := a.S.DB.Exec(ctx, `INSERT INTO gm_reviews(member_id,game_id,rating,comment) VALUES($1,$2,5,$3)`, memberID, gameID, strings.Repeat("a", 2000)); err != nil {
		t.Fatal("valid 2000-byte review rejected: ", err)
	}
}

// These helpers operate only on databaseAPI's disposable schema/search_path.
func dropIdentityConstraints(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	_, err := db.Exec(context.Background(), `ALTER TABLE gm_members DROP CONSTRAINT gm_members_username_key, DROP CONSTRAINT gm_members_username_identity_check, DROP CONSTRAINT gm_members_display_name_check; ALTER TABLE gm_reviews DROP CONSTRAINT gm_reviews_comment_check`)
	if err != nil {
		t.Fatal(err)
	}
}

func identityConstraintCount(t *testing.T, db *pgxpool.Pool) int {
	t.Helper()
	var count int
	err := db.QueryRow(context.Background(), `SELECT count(*) FROM pg_constraint WHERE (conrelid='gm_members'::regclass AND conname IN ('gm_members_username_key','gm_members_username_identity_check','gm_members_display_name_check')) OR (conrelid='gm_reviews'::regclass AND conname='gm_reviews_comment_check')`).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func TestIdentityMigrationPreservesRowsAndIsIdempotent(t *testing.T) {
	a, _ := databaseAPI(t)
	dropIdentityConstraints(t, a.S.DB)
	ctx := context.Background()
	var memberID, gameID, reviewID int64
	if err := a.S.DB.QueryRow(ctx, `INSERT INTO gm_members(email,username,password_hash,display_name) VALUES('existing@example.test','Existing','test-only-hash','Existing member') RETURNING id`).Scan(&memberID); err != nil {
		t.Fatal(err)
	}
	if err := a.S.DB.QueryRow(ctx, `INSERT INTO gm_games(title) VALUES('Existing game') RETURNING id`).Scan(&gameID); err != nil {
		t.Fatal(err)
	}
	if err := a.S.DB.QueryRow(ctx, `INSERT INTO gm_reviews(member_id,game_id,rating,comment) VALUES($1,$2,4,'Existing review remains unchanged') RETURNING id`, memberID, gameID).Scan(&reviewID); err != nil {
		t.Fatal(err)
	}
	migration, err := os.ReadFile(identityMigrationPath)
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, err = a.S.DB.Exec(ctx, string(migration)); err != nil {
			t.Fatal("migration failed: ", err)
		}
		if identityConstraintCount(t, a.S.DB) != 4 {
			t.Fatal("migration did not install all four constraints")
		}
		var count int
		err = a.S.DB.QueryRow(ctx, `SELECT count(*) FROM gm_members m JOIN gm_reviews r ON r.member_id=m.id JOIN gm_games g ON g.id=r.game_id WHERE m.id=$1 AND m.email='existing@example.test' AND m.username='Existing' AND m.display_name='Existing member' AND r.id=$2 AND r.rating=4 AND r.comment='Existing review remains unchanged' AND g.id=$3 AND g.title='Existing game'`, memberID, reviewID, gameID).Scan(&count)
		if err != nil || count != 1 {
			t.Fatal("migration changed or removed existing rows: ", err)
		}
	}
}

func TestIdentityMigrationRejectsInvalidRowsAtomically(t *testing.T) {
	for _, tc := range []struct {
		name, username, displayName, comment, code string
	}{
		{"duplicate username", "Existing", "Other member", "Existing valid review", "23505"},
		{"foreign email username", "victim@example.test", "Other member", "Existing valid review", "23514"},
		{"blank name", "Other", "   ", "Existing valid review", "23514"},
		{"short review", "Other", "Other member", "short", "23514"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := databaseAPI(t)
			dropIdentityConstraints(t, a.S.DB)
			ctx := context.Background()
			if _, err := a.S.DB.Exec(ctx, `INSERT INTO gm_members(email,username,password_hash,display_name) VALUES('existing@example.test','Existing','test-only-hash','Existing member')`); err != nil {
				t.Fatal(err)
			}
			var memberID, gameID int64
			if err := a.S.DB.QueryRow(ctx, `INSERT INTO gm_members(email,username,password_hash,display_name) VALUES('other@example.test',$1,'test-only-hash',$2) RETURNING id`, tc.username, tc.displayName).Scan(&memberID); err != nil {
				t.Fatal(err)
			}
			if err := a.S.DB.QueryRow(ctx, `INSERT INTO gm_games(title) VALUES('Existing game') RETURNING id`).Scan(&gameID); err != nil {
				t.Fatal(err)
			}
			if _, err := a.S.DB.Exec(ctx, `INSERT INTO gm_reviews(member_id,game_id,rating,comment) VALUES($1,$2,4,$3)`, memberID, gameID, tc.comment); err != nil {
				t.Fatal(err)
			}
			migration, err := os.ReadFile(identityMigrationPath)
			if err != nil {
				t.Fatal(err)
			}
			_, err = a.S.DB.Exec(ctx, string(migration))
			requireSQLState(t, err, tc.code)
			if identityConstraintCount(t, a.S.DB) != 0 {
				t.Fatal("failed migration left partially installed constraints")
			}
			var members, reviews, unchanged int
			err = a.S.DB.QueryRow(ctx, `SELECT (SELECT count(*) FROM gm_members),(SELECT count(*) FROM gm_reviews),(SELECT count(*) FROM gm_members m JOIN gm_reviews r ON r.member_id=m.id WHERE m.id=$1 AND m.username=$2 AND m.display_name=$3 AND r.comment=$4)`, memberID, tc.username, tc.displayName, tc.comment).Scan(&members, &reviews, &unchanged)
			if err != nil || members != 2 || reviews != 1 || unchanged != 1 {
				t.Fatal("failed migration changed or deleted existing rows: ", err)
			}
		})
	}
}
