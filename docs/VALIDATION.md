# Historical validation

The results below were recorded in the original monorepo. Evidence files remain there: https://github.com/nonthawat2548/GameMatch_ProjectWeb_BEFE/tree/main/docs/evidence . Current split-repo checks are in REPOSITORY_SPLIT.md.

# Validation record — 10 November 2026 (Asia/Bangkok)

## Latest popular catalog verification — 100 selected games

The user requested at least 100 popular/trending games. On 10 November 2026 the actual Go maintenance command `go run ./cmd/import-popular` verified 125 candidates and committed **100 eligible games** from Steam Most Played, supplemented by deduplicated Top Sellers TH entries. The live public API now returns **112 games**, retaining the previous 12 releases. Three software products discovered during validation (Wallpaper Engine, OBS Studio, Crosshair X) are hidden rather than deleted. There are exactly 100 published rows with popularity metadata. These are a manually refreshed selection, not claimed live global ranks/player counts. Existing descriptions (including empty ones) and publication are preserved on reimport.

- Source/import evidence: `evidence/steam-popular-import.txt`; verified live API snapshot `evidence/steam-popular-catalog.json`. Additive migration adds popularity_order/source/updated_at; no MovieHub data or user favorites were removed.
- Format, **13 React suites / 86 assertions**, production build, Go tidy/fmt/vet/build and **25 passing top-level Go tests** passed. One optional real-network New Releases test is explicitly skipped in this run. The new isolated-schema integration test proves deduplication, source ordering, preservation of Admin fields, incomplete-fetch rejection and rollback of ranking reset/upserts after a forced SQL failure. Existing API ownership/role/revocation/RAWG rollback tests remain passing.
- Hosted headless Chromium verified **all 112 actual CDN covers on all five catalog pages** at each of 1366/390/320px, no placeholders/page errors/overflow. It also verified search resets pagination, alphabetical sort, popular Home/source links and dates. `evidence/steam-browser/report.json` and nine updated screenshots; desktop and narrow Home were visually reviewed. This process used fresh isolated profiles and no active user screen.
- Hosted Guest/Member/Admin journeys passed all five groups against the new real catalog: registration/login return, persistent favorites, ownership/401/403, persisted preferences, Admin CRUD/publication/public 404, profile/password revocation/logout. Disposable QA accounts/game were cleaned up. `evidence/cloud-browser/report.json`. No RAWG key or new live RAWG import is claimed.
- Application source **74265de3896cacfc26e3c1e9c550d76d2acb5ff8**: Render `dep-dbpdg6avcj2c73d1o52g` verified **live**, Vercel `dpl_BBAtJbVw3tJo4cZEZmbP8KgTp7sH` verified **READY**. Public URLs remain https://gamematch-projectweb-befe.vercel.app and https://gamematch-api-hzv6.onrender.com. React fetches once and filters/sorts locally, showing 24 games per page. The Admin Steam HTTP action remains the separate 12-New-Releases import; the 100-game maintenance command is manual and server-side.

## Earlier Steam catalog / image verification

The user requested recent real games and images from Steam or another source. Added a separate guarded Go handler `POST /api/admin/sync/steam`, an additive three-column migration (steam_id unique, release_date, synced_at), and date/source-aware React cards/detail/Home. Real Go → Steam Store → Supabase import persisted **12 eligible New Releases**, with release dates supplied by Steam (9 November 2026 in this snapshot), images, categories and Store tags. The four exact original Demo rows were hidden rather than deleted; no user account, favorites or custom game was erased. A separate user-requested Admin promotion was completed and verified; the private email is not recorded in this public repo.

- Frontend format/check, **13 suites / 84 tests**, production build passed. New tests verify Steam image/source/date display and Admin Steam request/reload behavior. Existing RAWG compatibility assertions remain passing.
- Go tidy/fmt/vet/build and **23 passing top-level tests** passed with four isolated-schema Supabase integration tests. One additional real-network import test is explicitly skipped by default; it was run separately with opt-in and passed the actual import (12 entries). Steam checks cover upcoming/DLC/future/ambiguous dates, missing/changed tags, duplicate-safe updates, preserved publication/descriptions including intentional empty descriptions, source capsule selection, guest401/member403, and transaction rollback after a real SQL failure.
- Hosted image QA: `evidence/steam-browser/report.json` verifies **all 12 actual CDN covers load with naturalWidth >0**, no placeholders or uncaught page errors, source links/dates and no horizontal overflow at 1366/390/320px. Nine screenshots were captured and reviewed as local image files, without accessing the user's desktop/browser.
- Hosted full user journeys were repeated against the real Steam catalog. `evidence/cloud-browser/report.json` now includes an actual Admin Steam button fetch/import plus the existing Guest/Member/Admin/auth/persistence flows; temporary QA users/game were removed.
- Source snapshot: `evidence/steam-catalog.json`. Steam sync reads public Store endpoints, at most 20 candidates and 12 eligible games, manually on Admin request. This is a feed snapshot, not an exhaustive all-platform ranking or a scheduled updater. These endpoints are not a guaranteed versioned Steamworks contract; upstream/schema changes fail with 502 and keep the old catalog. Genre/category/Store tag values are mapped without inferring co-op from generic multiplayer. Descriptions are plain text. Per-game and page-level Steam credits/links are present; RAWG stays available as an optional separate import requiring a real key.

The following sections retain earlier snapshots; their older Demo-only/RAWG-key-pending and test counts do not describe the latest Steam catalog.

## Latest full-stack run — 10 November 2026

- Created **gamematch-db** (`mcwhgbiodyqotfgwzgkd`) in the user-selected Supabase organization, Singapore, after the user explicitly accepted the quoted **$0/month**. Applied `database/schema.sql` and the four explicitly fictional `Demo:` seed rows. MovieHub was not resumed or modified.
- Provisioned a trusted backend login with rights only to the GameMatch tables/sequences and RLS bypass. A separate QA role owns disposable test schemas. Verified connection through the Supabase session pooler with SSL; direct IPv6 endpoint was unavailable on this machine. Credentials live only in ignored local environment/cache files and provider environment settings, not source/evidence.
- RLS is enabled on all four public tables; `anon` and `authenticated` have no SELECT privilege. Supabase security advisor returned only four informational “RLS enabled, no policy” findings, consistent with the intentional backend-only access model. Reference: https://supabase.com/docs/guides/database/database-linter?lint=0008_rls_enabled_no_policy . No client access policy was added.
- **Go tidy/fmt/test/vet/build passed**, including **20 top-level test functions** and six upstream-response subcases. Three real Supabase integration tests create random private schemas from the actual schema.sql, run HTTP handlers against PostgreSQL, and drop only their own schemas. Verified two-account favorite isolation, persisted preferences, current DB role 403 checks, password/logout JWT invalidation, Admin CRUD/publication, unique/FK/role constraints and a mid-import SQL failure that rolls back earlier inserted rows. RAWG upstream responses in these tests are fixtures. Raw outputs: `evidence/backend-check.txt`, `evidence/backend-tests.txt`.
- Added shared Login/Register throttling (30 attempts/TCP peer/minute, bounded map, concurrent/expiry/spoofed-header tests), a five-connection DB pool, and strict RAWG JSON/body/page bounds. Proxy/NAT peers can share a limit; it is single-process and resets on restart.
- `govulncheck v1.9.0` initially detected reachable findings in older pgx/JWT/x-text. Updated within the existing libraries to pgx v5.9.2, JWT v5.2.2, x/crypto v0.52.0 and x/text v0.41.0; their requirements raise minimum Go to **1.25**, tested using **1.27.3**. Latest symbol scan: **0 affected vulnerabilities**, 0 vulnerable imported packages; four findings remain in uncalled modules. See `evidence/backend-vulnerabilities.txt`; the earlier JSON is retained as `backend-vulnerabilities-before.json`. No framework or persistence method changed.
- Frontend **format → 13 suites / 82 tests → production build passed**. Real latency exposed a Quiz race: user input could occur while the saved session was still restoring and be reset when the member mounted. Quiz now waits for session and preference restore, with a regression assertion. Combined evidence: `evidence/frontend-check.txt`.
- **Actual browser → Go → Supabase** verification passed in isolated headless Chromium: five scenario groups covering Guest Quiz/detail, UI registration/auto-login return, favorite reload, two-member API isolation/401/403, preference reload, Admin create/edit/hide/delete, profile refresh, password change, old JWT rejection, new-password login and logout. Four screenshots and sanitized request status evidence: `evidence/live-browser/report.json`. Disposable QA members/game were removed; original catalog preserved. Images are empty/fallback for Demo seed rows. This is real local Go + hosted DB, not a production deployment or live RAWG result.
- The separate mock browser suite still passed ten groups at desktop/mobile/narrow sizes after the fix. No active-screen control, visible browser window, reused user profile or CUA action was used.
- Git repository initialized locally and connected to the user's `https://github.com/nonthawat2548/GameMatch_ProjectWeb_BEFE`. The original initial README commit is retained. Deployment state is recorded in `DEPLOYMENT.md`; production UI and live RAWG remain unverified until recorded there. User did not have a RAWG API key, so no real import is claimed.
- Pushed reviewed source/evidence as commit `5744c83085a07125b061fa7b42f7c1cb0d2fbd5f`; staged scan found no generated credentials, .env, tools, node_modules or binaries. Created a separate free Render service in the selected Workspace, main/Go1.27.3/Singapore. Deploy `dep-dbpcf1gm7kps73ahhm90` became **live**, then real API tests at `https://gamematch-api-hzv6.onrender.com` verified DB health, Register/Login, 401/403, favorite ownership and logout revocation. Cloud QA accounts were cleaned up. Evidence: `evidence/render-live-api.json`. Vercel frontend access was still awaiting the provider team-consent grant at this checkpoint.
- After the user completed Vercel team consent, created a separate GameMatch project and production deployment: **READY**, `https://gamematch-projectweb-befe.vercel.app`, Node22/CRA/root frontend. Render CORS update `dep-dbpchuvg5rvs738he1k0` became **live** for the verified short production domain and team alias. Both releases use application source at `fda44e555d3e22c29ba685a82ca6eb2070dd6608`. Hosted browser verification then passed **five groups** through actual Vercel → Render → Supabase with no HTTP mocks or uncaught page errors; QA users/game cleaned up. Evidence and screenshots: `evidence/cloud-browser/`. Deep-route SPA responses and allowed/denied CORS passed in `evidence/production-smoke.json`. This supersedes earlier deployment-pending notes below. RAWG key/import, permanent team Admin email and the team's own presentation remain outstanding setup/content tasks, not verified deliverables.

## Earlier frontend phase — retained history

Latest frontend check was run with **`npm run check`**, covering format → 13 suites / 81 tests → production build; all passed. Combined raw output: `evidence/frontend-check.txt`; individual evidence files are sections of that same run. Home, catalog and About are separate routes; Home previews up to four actual API entries. Catalog now loads through fetch once per mount/retry and filters title/genre using React state, matching the course example's Movies pattern. Home tests cover loading, empty, failure/retry and preview navigation. Eight App checks cover the real React layers with mocked HTTP.

The user also supplied the lecturer's actual project announcement and grading image. `docs/RUBRIC.md` was updated to the group 13 / clip 7 split, including per-person explanation and evidence requirements. This does not certify live deploy/database marks.

Latest UI revision follows the user's MovieHub preference: light catalog layout, Thai navigation, compact image-led cards, real-genre filters and a collapsible score explanation. MovieHub Home/Navbar/MovieCard/Movies/index.css/App source files were read before implementing. React **13 suites / 81 tests passed**, including filtering on supplied game genres without another API call. Format check and production build passed. Isolated headless Chromium then verified the production build with mock HTTP/image fixtures at 1366/390/320px. It opened no windows, reused no user browser/profile and controlled no active screen. Details and limits are below.

- Read AGENTS.md, CODEX_PROMPT_TH.md, README and required handoff/API/rubric/validation documents. Inspected actual course repository files; details in REFERENCES.md.
- `npm ci --legacy-peer-deps`: completed after approval for npm cache/network access. Added pinned Prettier 3.6.2 and format scripts. Updated React Router within version 6 (6.30.6) and AJV within version 8 (8.20.0) following audit findings; retained CRA/React 19/Tailwind 3/native fetch.
- React latest run: **13 suites / 81 tests passed**; see `evidence/frontend-tests.txt`. Includes auth restore/login/logout and races, cross-tab session changes, registration recovery, UTF-8 byte limits, profile/password, favorites UI, Discover, Quiz, Detail navigation/fallbacks, Admin CRUD/publication/Sync/failures and malformed/cancelled JSON transport. Component tests use API/auth mocks. Eight App tests use real Router, Context and fetch helper with mocked HTTP responses; these are not Go/DB integration results, real multi-tab browser QA or production authorization evidence.
- Tests are separated into `frontend/src/tests/` (App test plus `api/`, `auth/`, `components/`, `pages/`) and `backend/tests/`, as requested. React 81 tests passed in the latest run. Go 8 tests, fmt/vet/build passed after the earlier test-folder move; backend code was not changed this turn. CRA shared setup remains `src/setupTests.js` by convention.
- Frontend `npm run build`: **Compiled successfully**; see `evidence/frontend-build.txt`. `npm run format:check`: passed. CRA dev server compiled and was available at `http://localhost:3000` at the time of this run.
- Downloaded a workspace-local Go toolchain from go.dev and checked its published SHA256; `go version` reported `go1.27.3 windows/amd64`. Ran `go mod tidy`, `go fmt ./...`, `go test ./...`, `go vet ./...`, `go build -o gamematch.exe .`: **passed**. Existing Go tests: 8 test functions; verbose output in `evidence/backend-tests.txt`. No new backend behavior was added while backend work is deferred.
- No Go on original PATH; local toolchain/cache excluded by .gitignore. Dependencies and lockfiles remain included. Actual secret environment variables were absent; source scan found no DATABASE_URL/JWT_SECRET/RAWG_API_KEY or unsafe HTML injection in frontend/src.
- Earlier Computer Use attempts failed or were stopped by the user, who disallowed active-screen control. This boundary remains in force. The later successful browser QA used an isolated headless process launched from the terminal, not Computer Use or the user's browser session.
- Headless Chromium **10 scenario groups passed**, with **34 screenshots** and no uncaught page JavaScript errors. Source: `frontend/tests/browser/verify.cjs`; replay instructions beside it. Results: `evidence/frontend-browser/report.json`. API and images are fixtures, not live Go/DB/RAWG or production authorization evidence.
- Supabase listing returned only inactive `movie-hub-db`. It was not resumed, modified or reused. No dedicated GameMatch credentials/schema/data/import were applied. Folder was **not a Git repository**, so no remote or Git history exists here to validate/push.
- Dependency audit remains a release limitation: latest explicit audit reports **36 findings (16 low, 7 moderate, 13 high)**, including remaining react-scripts tooling and react-router-dom advisory. Report saved in `evidence/npm-audit.json`. Router/AJV updates reduced some findings but did not clear the audit. No forced framework migration was attempted.
- Latest user instruction is to finish frontend before proceeding to backend. Frontend implementation, formatting, React assertions, production build and isolated browser QA are complete for this phase. Outstanding: login/register throttling, real DB integration, ownership/401/403/constraints, RAWG rollback and sync with user key, Vercel/Render deployments, production browser QA and recording by the two members. Project is **not yet certified ready to submit**.

## Frontend browser verification — isolated Chromium, mock HTTP

Story: Guest catalog/Quiz/detail → Login return path → Member Favorites/Preferences/Profile → Admin catalog maintenance, with the actual production React bundle and native fetch, but mocked HTTP data at the API boundary.

| Boundary | Result | Evidence / limit |
|---|---|---|
| UI rendering and responsive layout | Passed | Guest routes at 1366×900, 390×844, 320×720; member/admin forms; screenshots reviewed as local files; document width does not exceed viewport |
| UI → fetch → JSON → UI | Passed with fixtures | Search/genre filtering makes no additional catalog call; recommendations render five fixture results; login returns to detail; favorites and preferences survive reload of the mock store |
| Forms and role UX | Passed with fixtures | Profile refresh, password-change login notice/token cleanup, Admin add/edit/hide/delete, Sync error preserves existing rows |
| Native browser session events | Passed with mock API | Two tabs in a fresh context: logout in one hides favorite data and shows Login in the other, without reload |
| Loading / Empty / Error / retry | Passed | Controlled delayed/empty/503 catalog fixtures; no uncaught JavaScript page errors |
| Go → Supabase / real authorization / persistence | Unverified | Dedicated GameMatch DB credentials are still absent; mock tokens/store do not establish backend security or DB persistence |
| Production hosting and real RAWG data | Unverified | No deployment or live RAWG request was performed |

The browser test exposed a real layout defect: a valid 100-byte display name without spaces caused a 320px page to grow to 1039px. Navbar account links now wrap, and the regression check passes at 320px. The 404 page also provides working catalog/home links. Temporary Playwright/browser binaries are ignored under `.tools/` and `.cache/`; the application remains CRA + React 19 + Router 6 + Tailwind 3 + native fetch. This optional QA tool is not an application dependency.

## Frontend App integration checks (mock HTTP, actual React components)

1. Guest Discover → Quiz → 100/100 fixture score/reason → detail → login → JWT saved → favorite PUT with Bearer header → Favorites list → profile PUT and /api/me refresh → navbar name updated → logout and token removed: passed.
2. Protected Favorites URL with expired saved token → /api/me 401 → token removed → login form: passed.
3. Direct Admin URL as member → frontend 403 and no admin data fetch: passed. This verifies route UX only; backend member 403 and DB ownership still require live integration.
4. A protected API returns 401 while Favorites is open → current token removed → Context cleared → Login shown with no old favorite cards: passed.
5. Simulated storage event changes account A to B → old favorite cards hidden → /api/me fixture identifies B → B's empty favorites shown: passed. No real second browser tab or live DB is involved.
6. Protected URL receives temporary 503 during session restore → token and original URL retained → retry restores account without entering credentials again: passed.
7. Guest directly opens Favorites → Login → successful authentication returns to Favorites, rather than sending the member to the catalog: passed. The guest detail journey also returns directly to the selected game's detail after login.
8. Main navigation separates Home → all games → About, with the expected page headings and route paths: passed.

Login/register links retain the requested internal page. Allowed destinations are checked against actual app routes; tests reject external URLs, protocol-relative URLs, backslash variants, login loops and non-string state. Delayed signup/login/password-change continuations cannot overwrite or log out an account selected while the old request was pending.

Detail now separates the game image from facts/actions, with explicit genre/platform labels and tags. Profile forms use a shared two-column layout with a mobile single-column breakpoint. These layouts were verified in isolated headless Chromium and their screenshots reviewed; live deployment rendering remains unverified.

Additional regression checks: a delayed 401/restore/logout from an older session cannot erase or restore a newer account; failed login 401 and admin 403 do not clear an existing valid token. Detail navigation creates fresh per-game state, so late responses or favorite confirmations from the old game never appear on the new game.

Admin tests verify failure preservation and retry semantics: a successful write followed by list reload failure never repeats the write when retrying the list. RAWG sync responses are fixtures; no live import or DB rollback is claimed.

## Previous supplied starter record

The following section came with the starter. It describes that earlier creation session, not additional checks performed on the current machine.

## Actually run in this creation session
- Installed frontend dependencies with npm install --legacy-peer-deps, generated package-lock.json.
- Fixed CRA dependency resolution (ajv 8.17.1 pinned as a dev dependency).
- React: CI=true npm test -- --watchAll=false --runInBand: **2 suites / 6 assertions-based cases passed** (Quiz selected body/result, empty result, API failure; Guest redirect, Member admin denial, Admin allowed).
- Frontend npm run build: **Compiled successfully**. React production bundle ~71KB gzipped. Build output is omitted from source ZIP; rebuild after setting Production API URL.
- Go 1.23.6: go mod tidy, go fmt ./..., go test ./..., go vet ./..., go build: **passed**.
- Go tests include 8 test functions across scoring (3), JWT (3), unauthenticated guard (1), CORS (1).
- Checked ZIP integrity and source file inclusion; excludes node_modules, binaries, actual secrets, build output. Includes package-lock.json and go.sum.

## Outstanding integration work (starter baseline, still pending)
- No user Supabase DATABASE_URL or RAWG_API_KEY was available. No actual DB connection/schema application/import was performed.
- No live account was created and no resources were deployed or changed.
- No browser end-to-end QA was performed. FE tests mock API and auth context, so they are not proof of a real login/DB workflow.
- No DB integration tests yet: schema constraints, persistence, two-member ownership and role changes need verification against a dedicated GameMatch DB.
- No verified RAWG data response under the user's key; upstream tag coverage is variable.
- No production readiness or full rubric score is claimed.

## Acceptance scenario after environment setup
1. GET /healthz -> 200; schema applied; optional fixture seed present.
2. Guest Discover/detail/Quiz -> actual Go JSON; API unavailability -> Error, no matches -> Empty, slow API -> Loading.
3. Register accounts A and B; login A -> save game; login B -> cannot view/delete A favorite (all APIs scoped to own token).
4. Save preferences, refresh and return -> same selection restored.
5. Guest calls /api/me ->401; Member calls /api/admin/games or sync ->403 (test direct API, not only UI).
6. Promote one registered account through SQL; login -> Admin add/edit/hide/delete; hidden game excluded from public detail and Quiz.
7. Configure RAWG key -> Sync imports first-page metadata, repeated sync deduplicates by rawg_id; upstream failure leaves DB unchanged; published flag/admin description preserved.
8. Update profile -> navbar changes; wrong current password rejected; correct password invalidates old JWT; logout invalidates other sessions too.
9. Deploy FE+BE; set exact allowed origin; refresh /games/ID directly; ensure no secrets appear in built JS or Git history.
10. Run tests/build again after modifications and record actual outputs for video.
