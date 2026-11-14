# Go tests

The `tests` package tests exported auth, recommendation, middleware and handler APIs.
Production packages stay in `auth/`, `recommend/`, `mw/`, `handler/`, and `store/`.

Run from the repository root: `go test ./...` (or only this folder with `go test -v ./tests`).

Real PostgreSQL integration tests are in `database_test.go` and related regression/import test files. Set `GM_TEST_DATABASE_URL`
to a dedicated GameMatch test role with CREATE SCHEMA permission, then run
`go test -count=1 -v ./...`. Without that variable, the database tests are explicitly
skipped. A unit-only passing run is not DB evidence.

Each integration test creates a random `gm_test_<hex>` schema, applies the actual
`database/schema.sql`, and drops only that schema during cleanup. Public catalog
and member rows are not truncated. Never supply the MovieHub database.

Coverage includes two-member ownership, persisted preferences, current DB roles,
password/session invalidation, Admin visibility/CRUD, FK/unique/role constraints,
and a real SQL failure halfway through RAWG import followed by rollback. RAWG HTTP
responses in that test are controlled fixtures, not live RAWG data.

Rate-limit tests cover concurrent attempts, expiry, separate peers, bounded memory
and forged forwarding headers. Sync tests reject missing/oversized/malformed data.

Steam tests cover source mapping, released-game/date/tag checks, admin-only routes,
duplicate-safe updates, preservation of descriptions/visibility, and SQL rollback.
`TestLiveSteamImport` is **skipped by default**: setting GM_STEAM_IMPORT=1 and
DATABASE_URL explicitly imports real Steam data into that real database. It is
not a disposable-schema test; use it only for an intended GameMatch catalog update.

Regression coverage also checks username/email separation, duplicate usernames,
upgrade migration atomicity/idempotency and row preservation, helpful counts after
review edits, literal substring search, numeric game IDs, hidden-game visibility,
and favorite ownership. These database checks must run against a dedicated QA
database, never the production `DATABASE_URL`.

For meaningful application coverage when tests live in a separate package, run
`go test -coverpkg ./... -coverprofile coverage.out ./...` with the QA variable set.
The separate flag values also avoid PowerShell parsing issues with dotted paths.
