# GameMatch Backend

Go net/http, explicit handlers, JWT HS256 (2h), bcrypt, middleware, pgxpool and parameterized PostgreSQL SQL. React calls Go; Go connects to Supabase PostgreSQL. Supabase Auth is not used.

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

Go does not auto-load .env. Set terminal environment variables explicitly using the names in .env.example, or use Render dashboard settings:

```powershell
$env:DATABASE_URL = 'YOUR_SUPABASE_POSTGRES_CONNECTION_STRING'
$env:JWT_SECRET = 'REPLACE_WITH_AT_LEAST_32_RANDOM_CHARACTERS'
$env:ALLOWED_ORIGINS = 'http://localhost:3000'
$env:PORT = '8080'
$env:RAWG_API_KEY = ''
go run .
```

GET /healthz verifies DB connectivity. Keep credentials out of Git. Signup never accepts role; current DB role/token_version and favorite ownership are checked server-side. Logout/password changes invalidate all sessions. The bounded per-process login limiter can share quotas across proxy/NAT users.

## Render settings

- Repository: nonthawat2548/GameMatch_BE_Project, branch main.
- Root Directory: repository root (.), no backend subdirectory.
- Runtime: Go >=1.25; previous deployment used 1.27.3.
- Build: go build -o server .; Start: ./server; Health: /healthz.
- Preserve existing GameMatch DATABASE_URL/JWT_SECRET/RAWG_API_KEY in provider environment settings.
- ALLOWED_ORIGINS must contain the exact verified Vercel origin; Render supplies PORT.

Existing hosting used the old combined repo. This transfer does not reconnect Render, deploy a release or change the DB. See [split validation](REPOSITORY_SPLIT.md).

## Tests and API

[API contract](docs/API.md) and [test instructions](tests/README.md). Real DB tests require GM_TEST_DATABASE_URL for a dedicated QA role with CREATE SCHEMA permission; otherwise they are explicitly skipped. Passing unit tests alone are not DB evidence. Tests load database/schema.sql and remove only their disposable schemas.

Manual popularity import: go run ./cmd/import-popular, with an explicitly configured GameMatch DATABASE_URL. Admin Steam imports up to 12 New Releases; RAWG separately imports first-page 40 entries and requires a genuine server-side key. Imports preserve Admin descriptions/publication. Optional live import tests change real catalog data and are opt-in. Historical evidence and limits: [validation](docs/VALIDATION.md).
