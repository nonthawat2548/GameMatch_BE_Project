# Repository split ? 11 November 2026

This repo contains backend and database SQL at its root. See the root README for current setup and hosting root settings. All notes below are historical monorepo evidence; paths and old deployment links describe those earlier runs. Splitting source does not reconnect Vercel/Render or reapply a database schema.

# Handoff / architecture

Data path: React state/form -> api/backend.js fetch -> Go handler -> Guard JWT/current DB role -> store pgxpool SQL -> Supabase PostgreSQL -> JSON -> React state/UI.

## Latest catalog change — 10 November 2026

The user asked for real recent Steam/other games and images rather than empty Demo placeholders. Added `handler/steam.go`, guarded `POST /api/admin/sync/steam`, and steam_id/release_date/synced_at via an additive migration. Real Go import persisted 12 Steam New Releases with Store images/tags/release dates. The four original Demo rows were hidden by exact IDs/titles, not deleted. Main catalog sorts by release date; cards/detail link to Steam and display dates; Home identifies Steam releases; Admin has a manual Steam update button. RAWG remains optional and unchanged. Public Store endpoints are verified HTTP integrations, not an official stable/versioned Steamworks API guarantee; fail safely and retain old data on changes. Keep source attribution, never infer co-op from multiplayer, never invent style tags. Source snapshot: evidence/steam-catalog.json. The user's chosen registered account has already been promoted to Admin and verified; no need to ask for Admin email again.

## Latest backend handoff — 10 November 2026

The user asked to finish the project after the frontend phase. A new dedicated Supabase project `gamematch-db` (`mcwhgbiodyqotfgwzgkd`, Singapore) was created in `nonthawat2548's Org` after explicit $0/month confirmation; schema and four Demo seeds applied. Go connects through the verified session pooler with a scoped backend role, separate from the QA schema owner. `backend/.env` is configured locally and ignored. `scripts/Start-Backend.ps1` explicitly imports that file; Go itself still does not auto-load env files.

Latest checks: React 13 suites/82 tests + format/build, Go 20 top-level functions including three real Supabase integration tests + fmt/vet/build, five real browser-to-Go-to-Supabase groups, ten separate mock-browser groups. See VALIDATION and evidence. Temporary QA public accounts/game and isolated test schemas are cleaned up. No permanent/default Admin was created. A real latency bug in Quiz session restoration was fixed and regression-tested.

Backend now has a shared Login/Register limiter (30/TCP peer/min, bounded/concurrent tested, ignores forwarded headers), pool size five and strict RAWG response bounds. Limiter is per process and proxy/NAT users can share it. Dependencies were security-updated within pgx/JWT/bcrypt libraries, requiring Go 1.25+ (tested 1.27.3); govulncheck reports zero affected symbols, with four uncalled module findings remaining. `RawgClient` allows test transport injection; production URL stays fixed HTTPS.

User supplied repo: `nonthawat2548/GameMatch_ProjectWeb_BEFE`, main. Local Git now retains its initial commit. Selected Render workspace: `Workspace` (`tea-db866i8m7kps73d7qq60`). Vercel API initially returned team-grant 403 and the user was given a consent link; consult DEPLOYMENT for final status. User did not know/have a RAWG key: keep it empty and clearly report import unavailable, do not invent a key or claim fixture metadata is live. These facts supersede the older frontend-only status below.

Hosted path is now verified: https://gamematch-projectweb-befe.vercel.app → https://gamematch-api-hzv6.onrender.com → Supabase. Vercel production READY and Render live, both at application source fda44e555d3e22c29ba685a82ca6eb2070dd6608. Exact CORS domains are configured; five cloud-browser groups plus SPA deep-route checks passed. All temporary QA accounts/game were removed. The user was asked to register the real account and identify its email for Admin promotion; never create a default/password-in-code account. RAWG key remains genuinely missing. See DEPLOYMENT for resource IDs/evidence. Do not keep treating backend as deferred or deployments as absent based on the historical frontend notes.

## Current user preferences
- Do not control the user's active screen or browser windows. The user stopped Computer Use and explicitly disallowed this on 10 November 2026. Files/terminal and isolated headless software tests use no user browser/profile/desktop input. Screenshots from that isolated process have been reviewed as local image files.
- The user prefers the simpler MovieHub reference appearance. Current UI uses light surfaces, navy text, restrained purple accents, Thai navigation, compact landscape game cards and data-driven genre filter buttons. This replaces the earlier dark gradient/banner direction at the user's request. No frontend framework or backend behavior changed.

## Files to understand
- frontend/src/pages/Home.jsx, Discover.jsx, About.jsx: separate home/catalog/about routes using the reference's React state/effect/fetch approach. Catalog fetches on mount/retry and derives title/genre filtering locally; no per-keystroke request or daily browser cache.
- frontend/src/auth/AuthContext.jsx: restore JWT, login, logout, current user.
- frontend/src/components/ProtectedRoute.jsx: client route UX; not security boundary.
- frontend/src/pages/MatchQuiz.jsx: form, preference persistence, recommendation request.
- backend/handler/api.go: explicit routes, validation, scoped member operations and admin-only sync.
- backend/recommend/score.go: isolated deterministic rule function.
- backend/auth/token.go: strict JWT algorithm/issuer/expiry.
- backend/store/store.go: database CRUD. Handler file has inline simple queries aligned with starter scope; split by feature as needed.
- backend/mw/cors.go: allowlisted origins, preflight, no wildcard production access.

## Implemented / not yet certified
UI and handler paths are supplied for every basic feature. API/DB/browser integration needs a real configured DB and credentials. No hosting changes were made. See VALIDATION for actual checks.

## Next priorities for Codex
Current handoff, 10 November 2026: frontend is formatted with pinned Prettier (`npm run format`, `npm run format:check`). Tests live in `frontend/src/tests/`, optional isolated browser QA in `frontend/tests/browser/`, and Go tests in `backend/tests/`. Latest React run: 13 suites / 81 tests, including eight App integration checks with actual Router/Context/fetch helper but mocked HTTP. Headless Chromium checks the production build at 1366/390/320px, member/admin forms, native cross-tab logout and Loading/Empty/Error. API/image fixtures are not real Go/DB evidence. A valid 100-byte name without spaces exposed mobile navbar overflow (320px viewport, 1039px document); account links now wrap long names. Unknown routes now offer catalog/home recovery. Other coverage includes Admin CRUD/publication/Sync, form byte validation, client-side filtering, stale Quiz results, Favorites and RAWG source links. Registration recovery and Admin read-retry semantics are preserved. No DB/deployment resources were mutated. The latest user order is finish frontend first, then Go/DB; frontend implementation/isolated QA are complete for that handoff. Backend throttling, actual DB integration and deployment remain next.

Frontend session path: `auth/session.js` names the token key and a local session-expired event. `api/backend.js` dispatches that event only for a protected 401 matching the request's current stored token; late old requests and public login failures cannot erase a newer session. `AuthContext` listens for that event and native storage events, clears visible user data and revalidates account changes with /api/me. Shell remounts route content on member ID changes. Temporary restore outages retain the token/protected URL and offer retry. `refresh`/restore/logout compare captured tokens before changing account state. Native cross-tab logout passed in the isolated headless browser with a mock API. Go remains the authorization boundary; real DB revocation is unverified.

1. Reproduce builds/tests from clean install (`npm ci --legacy-peer-deps`, `go test ./...`). Dependencies are pinned; package-lock/go.sum included when generated.
2. Run new database schema/seed and real integration flow, including JWT invalidation and favorite isolation for 2 accounts.
3. Repeat browser checks against actual Go/DB and deployed URLs. Isolated frontend QA already covers desktop/mobile and controlled Loading/Empty/Error; do not count its mock store as DB persistence.
4. Add login/register throttling appropriate to Render proxy and avoid trusting arbitrary X-Forwarded-For. Add pagination to games if the catalog expands.
5. RAWG first-page sync is a deliberate starter scope; add bounded page selection, quota handling/backoff and import provenance if expanding. Missing tags cannot be assumed false. No price/session-time prediction.
6. Current detail description uses admin-authored text, RAWG list import does not fetch per-game full description. Do not claim all RAWG descriptions are imported.
7. Audit RAWG terms, images licensing and attribution links on all listing/result pages. Footer and detail links are a starting point, not legal certification.
8. Deploy correct FE/BE roots in new projects and verify direct nested-route refresh. Never overwrite MovieHub resources.
9. Capture real evidence and write 15-minute presentation. Actual two member names to be filled by team.

## Rubric scope choices
Admin has game CRUD + visibility + data sync. Member management and analytics dashboard are not implemented and are not required by the supplied basic Admin criterion. Do not claim those features exist. Use admin game list as console. Rule weights are 45/25/30 consistently (older chat's 50/35/15 and free-only/session filters were superseded because the upstream list does not verify those attributes).

## Security notes
User supplied the actual lecturer announcement and grading image: `docs/RUBRIC.md` now records the six group criteria totaling 13, team demo 2 and individual explanation/problem evidence 5. Read this before making submission-readiness claims. The latest instruction sets the order: finish frontend, then proceed to Go/DB. Do not reuse MovieHub resources or invent missing credentials.

Login uses `auth/returnPath.js` to return to the requested app page after authentication. ProtectedRoute and guest detail/quiz links set `state.from`; login/register transitions retain it. Only known internal routes are accepted. Do not accept arbitrary redirect URLs. Delayed login/signup responses compare the stored token before updating the active account; password-change completion checks the captured token before calling logout, so a previously selected account cannot log out a new account.

Supabase role that Go connects with must be trusted and authorized (RLS bypass e.g. provided postgres connection). Never solve denied access by giving anon access to password hashes. JWT contains id + token_version but the role is always looked up server-side. Changing role becomes effective at next API request. Change-password update checks old hash to avoid stale concurrent writes. Logout invalidates all tokens intentionally; temporary server outage still clears local token but cannot guarantee remote revocation. Before broad exposure: authentication rate limit, dependency audit and integration verification.
