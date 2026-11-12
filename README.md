# GameMatch Backend

Go net/http, explicit handlers, JWT HS256 (1h), bcrypt, middleware, pgxpool and parameterized PostgreSQL SQL. React calls Go; Go connects to Supabase PostgreSQL. Supabase Auth is not used.

Frontend: https://github.com/nonthawat2548/GameMatch_FE_Project

## Local setup

Use Go 1.25+ (original project tested with 1.27.3). Run from this repository root:

```powershell
go mod tidy
go fmt ./...
go test ./...
go vet ./...
go build -o server.exe .
```

For a NEW dedicated GameMatch database, apply database/schema.sql and optionally database/seed.sql (fictional Demo fixtures). Do not re-seed or replace the existing production database for this repository split. Never reuse MovieHub data. Go needs a trusted backend DB role; RLS blocks browser Data API access.

For local development, copy `.env.example` to `.env`, fill in your own credentials, then run `go run .` from the repository root. The server and `cmd/import-popular` automatically load this optional file. Existing process/hosting environment variables take precedence, including explicitly empty values. Invalid or unreadable `.env` files stop startup without printing credentials. `.env` is ignored by Git.

Admin imports now use **Steam only**. Set `STEAM_API_KEY` from https://steamcommunity.com/dev/apikey in the backend `.env` or hosting environment, then restart the backend. Never put the key in React environment variables or commit it. The Admin page checks `GET /api/admin/sync/status` for key presence only; this does not verify validity or quota.

`POST /api/admin/sync/steam` reads up to 12 game IDs from the official `IStoreService/GetAppList/v1/` using the key, then enriches them from Steam Store details and tag pages. It returns `imported`, `nextAppId`, and `hasMore`. Pass `?after=<nextAppId>` for the next batch. IDs are ordered by AppID, so this is a catalog import rather than a New Releases feed. Unavailable, unreleased, software, or unsuitable games are skipped; a batch can import zero games and still advance. The cursor advances in the UI only after a successful response. Reloading the Admin page starts a new import sequence; existing Steam IDs are updated without duplicates, preserving Admin descriptions and publication settings. Store metadata endpoints can change independently of the official API.

The older RAWG API endpoint is retained for compatibility, but the Admin page never calls it and no RAWG key is needed.

You can alternatively set terminal environment variables, or configure the hosting environment:

```powershell
$env:DATABASE_URL = 'YOUR_SUPABASE_POSTGRES_CONNECTION_STRING'
$env:JWT_SECRET = 'REPLACE_WITH_AT_LEAST_32_RANDOM_CHARACTERS'
$env:ALLOWED_ORIGINS = 'http://localhost:3000'
$env:PORT = '8080'
$env:STEAM_API_KEY = 'YOUR_STEAM_WEB_API_KEY'
go run .
```

GET /healthz verifies DB connectivity. Keep credentials out of Git. Signup never accepts role; current DB role/token_version and favorite ownership are checked server-side. Logout/password changes invalidate all sessions. The bounded per-process login limiter can share quotas across proxy/NAT users.

## Render settings

- Repository: nonthawat2548/GameMatch_BE_Project, branch main.
- Root Directory: repository root (.), no backend subdirectory.
- Runtime: Go >=1.25; previous deployment used 1.27.3.
- Build: go build -o server .; Start: ./server; Health: /healthz.
- Preserve existing GameMatch DATABASE_URL/JWT_SECRET; configure STEAM_API_KEY in provider environment settings.
- ALLOWED_ORIGINS must contain the exact verified Vercel origin; Render supplies PORT.

Existing hosting used the old combined repo. This transfer does not reconnect Render, deploy a release or change the DB. See [split validation](REPOSITORY_SPLIT.md).

## Tests and API

[API contract](docs/API.md) and [test instructions](tests/README.md). Real DB tests require GM_TEST_DATABASE_URL for a dedicated QA role with CREATE SCHEMA permission; otherwise they are explicitly skipped. Passing unit tests alone are not DB evidence. Tests load database/schema.sql and remove only their disposable schemas.

Manual popularity import: go run ./cmd/import-popular, with an explicitly configured GameMatch DATABASE_URL. Admin imports up to 12 Steam catalog candidates per request using STEAM_API_KEY; the popularity CLI remains a separate manual tool. Imports preserve Admin descriptions/publication. Optional live import tests change real catalog data and are opt-in. Historical evidence and limits: [validation](docs/VALIDATION.md).
