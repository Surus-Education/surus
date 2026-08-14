# PLAN

Task board. Agents read this on start, update it on finish. Handoff protocol lives in `CLAUDE.md` → "Agent handoff protocol".

Status legend: `todo` · `in progress` · `blocked` · `done`

Last updated: 2026-08-13 (four stacked PRs pushed; Vercel confirmed tracking `main`)

---

## Task 1 — Workflow setup

Goal: repeatable multi-agent workflow with scoped tooling and a streamlined Vercel deploy.

| #   | Subtask                              | Status | Owner agent |
| --- | ------------------------------------ | ------ | ----------- |
| 1.1 | Install MCP servers, scope per agent | done   | main        |
| 1.2 | Author custom agents                 | done   | main        |
| 1.3 | Scoped rules (`api/`, `web/`)        | done   | main        |
| 1.4 | Vercel production deploy             | in progress — repo side done, both projects confirmed tracking `main`, PRs pushed and awaiting merge | main |
| 1.5 | Clear 19 `no-explicit-any` lint errors, make lint gate blocking | done | frontend |

### 1.1 MCP servers — done

Main chat is restricted by *omission*, not by a disable list: only main-chat servers live in `.mcp.json`. Every other server is defined **inline** in the `mcpServers:` frontmatter of the agent that needs it, so it is never registered in the main session at all.

`.mcp.json` (main chat only):

| Server       | Transport | Auth              |
| ------------ | --------- | ----------------- |
| `context7`   | stdio     | none (free tier)  |
| `filesystem` | stdio     | none, scoped to `${CLAUDE_PROJECT_DIR:-.}` |
| `sentry`     | http      | OAuth on first use |

Agent-scoped (inline, invisible to main chat):

| Server                | Transport | Agents                        |
| --------------------- | --------- | ----------------------------- |
| `neon`                | http, OAuth | `backend`, `debug` (SELECT only) |
| `playwright`          | stdio     | `frontend`, `debug`           |
| `figma-desktop`       | http, localhost:3845 | `frontend`           |
| `chrome-devtools`     | stdio     | `debug`, `planner`            |
| `sequential-thinking` | stdio     | `planner`                     |

`context7` and `filesystem` are declared in `.mcp.json`, so agents that want them reference them **by name** rather than redefining them: `backend` and `frontend` take `context7` for library docs, `lookup` takes both.

No credentials are committed — Neon and Sentry use OAuth, Figma is localhost, the rest are unauthenticated. `.mcp.json` is safe to commit.

First run will prompt to approve the project-scoped servers, and Sentry/Neon will each open a browser for OAuth. `figma-desktop` only responds when the Figma desktop app is open with Dev Mode + desktop MCP server enabled.

### 1.2 Custom agents — done

Five agents in `.claude/agents/`. Each carries its own MCP servers, a tools allowlist, and a pinned model.

| Agent      | Model               | Scope                  | Writes            | MCP |
| ---------- | ------------------- | ---------------------- | ----------------- | --- |
| `backend`  | `claude-sonnet-5`   | `api/` only            | yes               | neon, context7 |
| `frontend` | `claude-sonnet-5`   | `web/` only            | yes               | playwright, figma-desktop, context7 |
| `debug`    | `claude-sonnet-5`   | reproduction           | no (reports only) | chrome-devtools, playwright, neon |
| `planner`  | `claude-opus-4-8`   | `PLAN.md` only         | `PLAN.md` only    | chrome-devtools, sequential-thinking |
| `lookup`   | `claude-haiku-4-5`  | read-only search       | no                | context7, filesystem |

`deploy` agent was dropped — see 1.4.

Model notes: `claude-opus-4-8` is a legacy model, priced the same as Opus 5 ($5/$25 per MTok) with an older knowledge cutoff (Jan 2026 vs May 2026) — pinned deliberately, not by default. There is no Haiku 5; `claude-haiku-4-5` is the current Haiku.

`debug` has Neon but **`SELECT` only** — it has no file-write tools, but that restriction does not cover the database, so the constraint is enforced in its prompt. Schema and data changes belong to `backend`.

Cross-boundary rule: `backend` and `frontend` may not edit each other's directory. Work that spans both becomes two subtasks with a stated dependency.

### 1.3 Scoped rules — done

- `api/CLAUDE.md` — package layout, sqlc workflow, error contract, auth + cross-origin cookie rules, domain model, container constraints.
- `web/CLAUDE.md` — route groups, the two fetch layers, `lib/types.ts` mirroring, UI conventions, API origin rules.
- Root `CLAUDE.md` trimmed to what spans both: handoff protocol, agent roster, commands, deploy topology, branch facts. The per-directory detail was moved out rather than duplicated.

### 1.4 Vercel deploy

Two Vercel projects against this one repo:

- **web** — root directory `web/`, standard Next.js build.
- **api** — root directory `api/`, container deploy via `api/Dockerfile.vercel`. Vercel detects `Dockerfile.vercel` at the project root, builds it, pushes to Vercel Container Registry, serves from a Fluid-compute Function.

**Decided: production is the only usable deployed environment. All testing happens locally.** Merging to `main` deploys straight to prod.

**Corrected 2026-08-13.** Earlier revisions of this section said "no preview environments" and that previews were "scrapped." That was wrong as a statement of fact — a PR deployment check on PR 2 proved Vercel preview deployments are enabled and running. The *reasoning* held; the conclusion was overstated.

Accurate position: previews exist and are **build verification only, not a test environment.** Rotating per-branch hostnames can never complete a Google OAuth login, so a preview cannot get past sign-in, which puts nearly the whole app out of reach on one. A green preview means "it compiles and deploys." Nothing else.

Consequences unchanged: functional testing is local against a Neon branch, and production is the first place the deployed cross-origin behavior is genuinely exercised. Leave previews on — they are a free extra build signal and cost nothing to ignore.

Consequences of that choice, stated plainly so nobody re-derives them later:

- There is **no staging gate**. Every push to the production branch is live. The GitHub Actions build gate is the only automated check between a push and production.
- Google OAuth needs a **stable, pre-registered** redirect URL — the production API's `/auth/google/callback`. This is the constraint that makes previews unusable for testing. (Earlier revisions said "exactly **one**." Too strong — see the correction under 2.1. Google accepts a list; the real constraint is that every entry must be known ahead of time, which rotating preview hostnames can never satisfy.)
- `CORS_ALLOWED_ORIGINS` needs only the exact production web origin. The wildcard matching in `middleware.CORS` stays in the code (guarded and documented) but should be left unused — an exact origin is strictly safer.
- Confirm which branch Vercel treats as the Production Branch. Current work sits on `vercel`, while `main` is the repo's default.

**Deployment stays on Vercel's native Git integration, not GitHub Actions deploy hooks.** Deploy Hooks build "latest commit on a configured branch" with no tie to a specific push, so driving them from Actions duplicates the built-in behavior and loses the commit association.

To configure, in Vercel project settings rather than CI:

- **Root Directory** per project (`web/`, `api/`).
- **Ignored Build Step** so the web project skips builds when only `api/` changed, and vice versa — still worth having, since both projects share one repo.

GitHub Actions is a **CI gate, not a deploy trigger**: `go build ./... && go vet ./...` for the API and `pnpm build` for the web.

Constraints the container runtime imposes, not yet handled in code:

- **Scale-to-zero after 5 min idle.** `api/internal/jobs/linkrot.go` is an in-process 24h ticker — it will effectively never fire in production. Must become an HTTP endpoint driven by Vercel Cron.
- **Stateless, no durable local storage.** Nothing may rely on process memory persisting between requests.
- **Neon pooled connection string required** — cold starts churn connections; the direct URL will exhaust limits.
- **CORS + cookies.** Auth is a JWT in an `access_token` cookie and the two projects sit on different origins. Needs `SameSite=None; Secure` plus a CORS allowlist containing the exact production web origin.
- **Google OAuth redirect URLs** are registered at root (`/auth/google/start`, `/auth/google/callback`) with no `/v1` prefix. The production entry is one fixed URL; local development adds a second, equally stable one (see the correction under 2.1). Preview hostnames can never be registered, which is why previews cannot complete a login.

Deliverable: pushing to the production branch deploys both projects to production, and a logged-in user can use the app there. All verification before that push happens locally.

#### Repo side — done

- `web/vercel.json`, `api/vercel.json` — `ignoreCommand` compares against `VERCEL_GIT_PREVIOUS_SHA` (falling back to `HEAD^`) scoped to the project's own root directory, so each project skips builds for changes that don't touch it.
- `.github/workflows/ci.yml` — build/lint gate only, no deploy trigger. `go build` + `go vet` for the API, `pnpm build` for the web, `pnpm lint` blocking (1.5 cleared the pre-existing errors and removed `continue-on-error`).
- `api/.env.example` — documents `NEXT_PUBLIC_APP_URL`, `CORS_ALLOWED_ORIGINS`, and what `APP_ENV` now controls.
- Cross-origin auth fixes in the API (see Completed) — without these production cannot work at all, previews or no previews.

#### Dashboard side

1. ~~Confirm which branch is the **Production Branch**~~ — **confirmed 2026-08-13: both projects already track `main`.** No change needed. `main` is therefore the merge target for all work, and merging deploys to production.
2. **Root Directory** per project: web project → `web/`, api project → `api/`.
3. Confirm the api project detects `Dockerfile.vercel` and builds a container.
4. Production env vars:
   - api: `DATABASE_URL` (Neon **pooled**), `JWT_SECRET`, `APP_ENV=production`, `NEXT_PUBLIC_APP_URL` (production web origin), `CORS_ALLOWED_ORIGINS` (exact production web origin), `GOOGLE_REDIRECT_URL`, Google/Postmark/R2 secrets.
   - web: `NEXT_PUBLIC_API_URL` pointing at the production api origin, **including the `/v1` suffix**. Strong indirect evidence it is already correct: the pre-fix sign-in link only resolved if this value ended in `/v1`, and production OAuth did reach Google's consent screen.
5. Register the production api `/auth/google/callback` with Google. Localhost is registered alongside it as of 2026-08-13 — see the OAuth subsection under 2.1.

### Shipping — four stacked PRs pushed 2026-08-13

Branches are on `origin`, PRs not yet opened. Merge **in order**; each is based on the one above it,
not on `main`.

| # | Branch | Base | Contents |
| - | ------ | ---- | -------- |
| 1 | `chore/dev-tooling-and-docs` | `main` | `pnpm dev` startup, CI gate, README, CLAUDE.md files, PLAN.md, vercel.json |
| 2 | `refactor/web-strict-types` | PR 1 | task 1.5 — `any` removal, lint made blocking |
| 3 | `fix/auth-session-bugs` | PR 2 | tasks 2.8, 2.10 (hash), 2.11, plus the 1.4 CORS/cookie fixes |
| 4 | `feat/cross-origin-session-bridge` | PR 3 | task 2.7a–c, the Option B handoff |

Each branch was verified independently: `go build`/`go vet` pass on all four, `pnpm build` on all
four, `pnpm lint` clean from PR 2 onward. PR 4's tip is byte-identical to the pre-split snapshot, so
nothing was lost in the split. A local `wip/all-changes` branch holds that snapshot until the PRs
merge.

The split reconstructs real history rather than the end state: PR 1 lands CI with lint
`continue-on-error` **because lint was not clean yet**, and PR 2 removes it. Don't "fix" that
apparent inconsistency later — it is deliberate.

`.claude/` and `.mcp.json` are now gitignored (agent definitions and MCP wiring are one person's
local setup). The `CLAUDE.md` files and `PLAN.md` are tracked. The agent roster in the root
`CLAUDE.md` documents a division of labour, not files present in a fresh clone.

#### Production schema — verify before PR 4 merges

The user reports the schema was applied to the production Neon branch at first deploy.

**That cannot cover `handoff_codes`.** That table was created today as part of 2.7a and did not exist
at any earlier deploy, so a schema applied "when I first deployed" predates it. Either it was applied
separately today, or it is missing.

This is worth ten seconds of checking because the failure is total and silent:
`CreateHandoffCode` errors, `GoogleCallbackRedirect` takes its error path, and **every production
login lands on `/?error=auth_failed`** with nothing obviously wrong in the logs.

```sql
SELECT to_regclass('public.handoff_codes');  -- NULL means missing
```

If NULL, apply the DDL from 2.7a **before** PR 4 merges — merging deploys straight to production,
and there is no staging gate to catch it.

---

## Task 2 — Get implemented features working

Goal: course list after login, course creation, and course elements all functional.

**Correction to the original premise:** the features are not sitting on an unmerged branch. Verified:

- `list` branch is identical to `main` (`git diff main list` → empty).
- `origin/eric/course-create` (`f4473ed`) is the same change as `main`'s `a5e4da7`, just a different commit hash off `b4372eb`.
- `vercel` = `main` + two Dockerfile commits.

All feature code is already in `main`: `web/app/(auth)/dashboard`, `(auth)/library`, `(auth)/courses/new`, `(auth)/courses/[courseId]/edit`, and `api/internal/{course,lesson,quiz,fork,auth,user,report}`. So this is debugging, not merging.

| #   | Subtask                                  | Status | Owner agent |
| --- | ---------------------------------------- | ------ | ----------- |
| 2.1 | Local dev against Neon dev branch        | done — env fixed, both services running, Google localhost redirect URI registered | you (env) |
| 2.2 | Reconcile with the author's working machine | done — env drift, not code | main |
| 2.3 | Reproduce failures, capture exact errors | done — no reproduction needed; root causes were env, not code | main |
| 2.4 | Course list after login                  | done — verified in browser | main |
| 2.5 | Course creation                          | done — verified in browser | main |
| 2.6 | Course elements (lesson tree + subtypes) | todo — **not** covered by the 2026-08-13 pass | tbd |
| 2.7 | Fix cross-domain session cookie (prod login) | in progress — 2.7a/b/c done, 2.7d open | backend + frontend |
| 2.7a | API handoff endpoint: mint a single-use code in `GoogleCallbackRedirect`, redirect to the web callback with it, add `POST /v1/auth/handoff` to exchange it for a token pair | done | backend |
| 2.7b | Web callback route (`app/auth/callback/route.ts`) exchanges the handoff code and sets both token cookies on the web origin | done | frontend |
| 2.7c | Logout only clears the API-origin cookie jar; the web-origin copy set by 2.7b survives | done | frontend |
| 2.7d | Logout still doesn't clear the **API-origin** cookie — 2.7c inverted the problem rather than closing it | todo | backend, then frontend |

### 2.7d — the other half of the logout gap

2.7c's route handler is correct in what it does, but its original code comment overstated it (comment
corrected in `web/hooks/useAuth.ts`). The API call it makes is issued **from the Next server**, so the
API's `ClearTokenCookies` deletion headers land on the server-to-server response and are discarded.
The browser never sees them, and the API-origin `access_token` survives.

Net effect after clicking sign out:

| Jar | Cleared? | Consequence |
| --- | -------- | ----------- |
| web origin | yes | Server components see no session; `(auth)` correctly redirects. UI looks logged out |
| API origin | **no** | `apiFetch` from client components still authenticates for up to 900s |

Refresh tokens and handoff codes *are* revoked in the database, so the session cannot be extended —
but `access_token` is a stateless JWT and stays valid until its own expiry. So the window is bounded
at 15 minutes, not indefinite. Still wrong: on a shared machine the next person can hit authenticated
API endpoints as the previous user for that window.

Fix requires a **backend** change first — the API has no GET logout, and you cannot navigate a browser
to a POST. Add a GET (e.g. `/auth/logout/redirect`, mounted at root alongside the OAuth redirect pair,
not under `/v1`) that calls `ClearTokenCookies` and 302s to a web URL. Then `frontend` chains it: the
web logout route clears its own jar and redirects the browser through that API endpoint, which clears
the API jar as a genuine top-level navigation and bounces back to `/`.

Guard the return URL against open redirect — send the browser to `NEXT_PUBLIC_APP_URL` rather than to
an arbitrary `?next=` parameter.

**This is Option B debt, and Option A deletes it.** With one shared parent domain there is one jar,
one deletion, and no chaining. Don't over-invest here.
| 2.8 | Implement `RefreshSession` — currently a stub | done | backend |
| 2.9 | Fix hardcoded `NEXT_PUBLIC_API_URL` / `/v1` suffix, local + Vercel | todo | you (env) |
| 2.10 | Magic-link login: hash mismatch makes every link invalid, and no email is ever sent | in progress — hash fix done, Postmark wiring still pending, see note | backend |
| 2.11 | Verify OAuth `state` in the callback (CSRF), fix its cookie attrs | done | backend |
| 2.12 | Delete dead `POST /v1/auth/google/callback` + unrouted `GoogleStart` | needs your OK | backend |

### 2.0 Root causes found by inspection (2026-08-13)

Three separate breaks. None is a code bug in the features themselves.

**(a) `NEXT_PUBLIC_API_URL` has no `/v1` and points at prod.** `web/.env.local` reads
`https://surus-1.vercel.app/`. All four fetch layers build URLs as `` `${API_URL}${path}` ``
(`web/lib/api/client.ts:21`, `lib/api/server/{courses,users}.ts`, `lib/auth/server.ts`) and every
API route except the two OAuth redirects is mounted under `/v1` (`api/cmd/server/main.go:98`).
Result: `https://surus-1.vercel.app//courses` → 404 on every course call. Explains 2.4, 2.5, 2.6
in one line. Local dev also never touches the local API at all.

**Check the Vercel web project's own `NEXT_PUBLIC_API_URL` too** — it is a separate value set in
the dashboard, and if it also lacks `/v1`, production has the identical 404.

**(b) The session cookie is set on a host the web app cannot read.** This is why deployed Google
login opens the consent screen and then silently lands you back on the landing page.

`GoogleCallbackRedirect` (`api/internal/auth/handler.go:179`) calls `SetTokenCookies`, so
`access_token` is scoped to the **API** host (`surus-1.vercel.app`). It then redirects to
`frontendURL + "/dashboard"` on `surus-beta.vercel.app`. There, `(auth)/layout.tsx` calls
`getServerSession()`, which does `cookieStore.get("access_token")` (`web/lib/auth/server.ts:8`) —
reading cookies for the **web** host. The cookie isn't there, so it returns `null` and the layout
`redirect("/")`s. The login succeeded server-side; the frontend can never observe it.

`SameSite=None; Secure` (fixed in 1.4) is necessary but not sufficient: it governs whether the
*browser* attaches the cookie to cross-site requests. `getServerSession` runs on the Next server
and reads its own request's cookie jar, which the API's cookie was never part of.

A `Domain` attribute cannot fix this — `vercel.app` is on the Public Suffix List, so
`Domain=.vercel.app` is rejected by every browser. See Open decision 2.

Works locally only because `localhost:8080` and `localhost:3000` are the same host; cookies ignore
port. That, plus (a), is the whole of "works on the author's machine."

**(c) `RefreshSession` is an unimplemented stub.** `api/internal/auth/service.go:218-223` returns
`unauthenticated` unconditionally — the body is a comment and a hardcoded error. `access_token`
has `MaxAge: 900`, so every session dies 15 minutes after login in every environment and cannot
be renewed. The `Path` fix in 1.4 made the cookie reach an endpoint that does nothing.
`GetRefreshTokenByHash`-style lookup doesn't exist yet; the comment says the intent was to iterate
non-revoked tokens and bcrypt-compare. Needs a real query.

**(d) Magic-link login cannot work either — same bug class as (c).** Found 2026-08-13 while
verifying 2.8. `GenerateMagicLinkToken` (`api/internal/auth/jwt.go:60`) stores a **bcrypt** hash,
which is salted and therefore different on every call. `VerifyMagicLink`
(`api/internal/auth/service.go`) then calls `GetMagicLinkTokenByHash(ctx, rawToken)` — passing the
**raw** token into a query whose predicate is `WHERE token_hash = $1`
(`api/db/queries/auth.sql:16-19`). A bcrypt digest never equals its plaintext, so the lookup always
misses and every magic link is rejected as "Invalid or expired token". The author's comment about
brute-force iteration acknowledges the mismatch without resolving it.

The fix is exactly what 2.8 did for refresh tokens: hash deterministically with SHA-256 (the raw
token is already 32 bytes of `crypto/rand`, so a salt buys nothing and only blocks the indexed
lookup), then look the row up directly. `HashRefreshToken` in `jwt.go` is the pattern to follow.

Second, independent defect in the same flow: `MagicLinkRequest`
(`api/internal/auth/handler.go:59-82`) never sends the email. It generates and stores the token,
discards the raw value into `_`, and returns 204 with a comment reading "In production, send email
via Postmark here". `POSTMARK_API_KEY` and `POSTMARK_FROM_EMAIL` are configured in `.env.example`
but nothing reads them. So even after the hash fix, no user can receive a link. Tracked as 2.10.

Lower severity, same area: `oauth_state` is generated and set but **never verified** —
`GoogleCallbackRedirect:165` reads the `state` param into `_`. That is a CSRF hole, not a cause of
the current symptoms.

**Correction (2026-08-13):** an earlier revision of this section, and of the 2.11 Completed entry,
claimed the `oauth_state` cookie's `SameSite=Lax` without `Secure` meant it "is dropped on the
cross-site return from Google." **That is wrong.** `SameSite=Lax` cookies *are* sent on cross-site
**top-level GET navigations**, and Google's redirect back to `/auth/google/callback` is exactly
that — this is why the Lax-state-cookie pattern is the standard OAuth recommendation. A cookie
without `Secure` is also still sent over HTTPS; `Secure` only *withholds* it over plain HTTP. So
the old cookie arrived fine. The only real defect in 2.11 was that `state` was never checked.

Where `SameSite` genuinely matters is the **session** cookies (2.0 (b) and 1.4): those are read by
`fetch`/XHR, not top-level navigation, and Lax does withhold them there. Don't generalize the
session-cookie rule to every cookie in the codebase — the transport differs.

Sharing `CookieAttrs()` for `oauth_state` is still fine and is what shipped, because the JSON
`GoogleStart`/`GoogleCallback` pair reaches the callback via `fetch`, where Lax would withhold it.
Correct outcome, wrong stated reason.

### 2.1 Local dev setup — decided: Neon dev branch

Neon is already provisioned and the current half-working build is deployed against it. Drop Docker Postgres for day-to-day work: point `api/.env` `DATABASE_URL` at a Neon branch (copy-on-write fork of prod, disposable), so debugging runs against the exact schema and data prod has.

**Confirmed 2026-08-13: schema and rows exist on the Neon production branch.** They need forking to a
dev branch for local work — do not point local `DATABASE_URL` at production.

Current `api/.env` is wrong in three ways and must be fixed before anything else runs:

| Key | Current | Should be |
| --- | ------- | --------- |
| `DATABASE_URL` | `…@localhost:5432/…` | Neon **pooled** URL for a dev branch forked off prod |
| `CORS_ALLOWED_ORIGINS` | absent | `http://localhost:3000` |
| `NEXT_PUBLIC_APP_URL` | absent | `http://localhost:3000` |

The last two were added to `.env.example` in 1.4 and never copied across. `main.go:59` always includes
`frontendURL` in the allowlist and defaults it to `http://localhost:3000`, so CORS happens to survive
locally — but the value is doing real work in production and should be set explicitly in both.

Docker is not running on this machine (`docker ps` → daemon not found), so the current
`localhost:5432` URL means `pool.Ping` fails and the API exits at `main.go:44` before serving a single
request. `web/.env.local` must go back to `http://localhost:8080/v1` or the frontend keeps talking to
production while you debug locally.

**Resolved 2026-08-13.** All three env files are correct and both services run. Verified by probe:
`http://localhost:8080/health` → 200 (so the API cleared `pool.Ping` and the Neon connection is
good), `http://localhost:3000/` → 200. A ping failure reported against `user=surus
database=surus_dev` after the edit is a **stale shell** holding the old `.env` — those are the
`.env.example` defaults, not the current file. Open a new terminal rather than re-debugging the URL.

#### Google OAuth: localhost must be registered too

`GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback` is sent verbatim as `redirect_uri`
by `Service.GoogleAuthURL`. Google rejects any value not pre-registered on the OAuth client, so a
fresh machine gets `redirect_uri_mismatch` until this exact string is added under **Authorized
redirect URIs** on the client (`439866506014…`) in Google Cloud Console → APIs & Services →
Credentials.

Three points that cost time if missed:

- The URI points at the **API** (`:8080`), not the web app (`:3000`), and carries no `/v1` — the
  OAuth redirect pair is mounted at root (`api/cmd/server/main.go:95-96`).
- Plain `http://` is valid here. Google makes an explicit exception to its HTTPS requirement for
  `localhost`, so no tunnel or proxy is needed.
- Adding it is **additive** — do not remove the production URI. This is why the "exactly one
  redirect URL" phrasing in 1.4 was wrong: Google accepts a list, and the previews reasoning rested
  on hostnames being *unpredictable*, not on the list being length one. Localhost is stable, so it
  registers once and stays.

**Authorized JavaScript origins** stays empty — this is a server-side redirect flow, not a browser
JS flow. Nothing belongs there.

Also needs fixing: `make up` is dead on Windows. It runs `nohup sh -c 'cd api && go run ./cmd/server'`; `make` hands that to a Unix shell (the log path `/mnt/c/nvm4w/nodejs/pnpm` indicates WSL) whose PATH has no Windows `go` or `node`:

```
.api.log → sh: 1: go: not found
.web.log → /mnt/c/nvm4w/nodejs/pnpm: 15: exec: node: not found
```

Fix the launcher, not the app. Two plain terminals (`cd api; go run ./cmd/server` and `cd web; pnpm dev`) is an acceptable interim.

### 2.2 Reconcile with the author's working machine — done

**Answer: environment drift, not code.** Nothing needs merging and nothing needs unstashing. The
five hypotheses below were the pre-investigation list; the actual causes are 2.0 (a) and (b) —
a `NEXT_PUBLIC_API_URL` missing its `/v1` and pointing at production, and a session cookie set on
a host the web app cannot read. Both are invisible to git because `.env` and `.env.local` are
gitignored, which is exactly why "it works on my machine" held.

The original hypothesis list, kept because it's still the right order to work a future report of
this shape:

1. **Empty or unmigrated database.** There is no migration runner, so the schema is applied by hand. A Neon branch that never had `001_initial_schema.sql` applied, or that has the schema but no rows, renders exactly as "course display doesn't work." — *Ruled out: prod branch has schema and rows.*
2. **Visibility and ownership.** `GetCourseByID` takes a `ViewerID`; private courses 404 for everyone but their owner. If the author's rows are owned by the author's user account and you're logged in as someone else, private courses correctly disappear. — *Still worth checking during 2.3 once requests actually reach the API.*
3. **Uncommitted work on the author's machine.** — *Ruled out: all feature code is in `main`.*
4. **Env differences.** `NEXT_PUBLIC_API_URL`, `DATABASE_URL`, `APP_ENV`. — **This was it.**
5. **Different database entirely.** — *Not the cause, but this machine did point at a dead `localhost:5432`.*

### 2.3 Reproduce — done, and it never needed a reproduction

The original framing assumed a runtime mystery requiring browser-and-log forensics. It wasn't. Every
root cause was found by reading code and env files (2.0 (a)–(d)), and the fix was environmental plus
four real auth bugs. No `debug` agent was ever dispatched.

Worth carrying forward: the cheap explanations — env drift, a database that isn't there, a URL
missing a path prefix — were the entire answer. The instinct to instrument the running system first
would have cost hours and found nothing, because the API wasn't booting.

### 2.4 / 2.5 — verified working 2026-08-13

Confirmed by the user against the Neon dev branch, in a browser:

- Google sign-in completes and lands on the dashboard.
- Courses can be created.
- Created courses appear in the list.
- Course search works.

This is the first time any of this session's work ran rather than compiled. It exercises, end to end:
the Option B handoff (2.7a + 2.7b), OAuth `state` verification (2.11), the CORS allowlist and cookie
`SameSite`/`Secure`/`Path` fixes from 1.4, and the `NEXT_PUBLIC_API_URL` correction (2.9).

**Still unverified at runtime, despite the successful pass:**

| Task | Why the pass didn't cover it |
| ---- | ---------------------------- |
| 2.6 | Lesson tree and the video/page/quiz subtypes were not exercised. Do not infer from "courses work" |
| 2.8 | Refresh rotation needs a deliberate wait past the 900s access-token TTL |
| 2.10 | Magic-link hash is fixed but no email is ever sent, so the flow can't be walked |
| 2.7d | Logout's API-origin cookie gap — known, accepted, not yet closed |

### Known broken

| Symptom | Where | Evidence | Root cause |
| ------- | ----- | -------- | ---------- |
| ~~`make up` starts nothing~~ **RESOLVED** | local | `.api.log`, `.web.log` | Makefile shell/PATH mismatch on Windows. Replaced by a root `package.json` running both services through `concurrently`, which spawns processes directly instead of via `nohup sh -c`. `make up` is now an alias for `pnpm dev` |
| Every authenticated request fails when deployed | api | code inspection | `access_token` was `SameSite=Lax` — browsers withhold it on cross-site requests, and web/api are on different origins. **Fixed in 1.4, confirmed by the working 2026-08-13 login.** |
| Session never refreshes; 401s become permanent | api | code inspection | `refresh_token` was scoped `Path=/auth/refresh` but the route is `/v1/auth/refresh`, so the cookie was never sent. **Fixed in 1.4.** The path is confirmed correct; whether refresh itself works still needs the >900s wait (2.8). |
| Logout may not clear session | api | code inspection | `ClearTokenCookies` omitted `Secure`/`SameSite`, so the deletion cookie didn't match the original and wrote a second cookie instead of overwriting. **Fixed in 1.4, not yet exercised** — logout was not part of the 2026-08-13 pass, and 2.7d is a separate open gap. |
| CORS rejects any origin but one exact URL | api | code inspection | Allowlist was a single hardcoded origin with no way to add others. **Fixed in 1.4, confirmed** — the 2026-08-13 login and course calls crossed origins successfully. |
| Course display works for the author, not here | local | 2.2 | **Env drift.** `web/.env.local` `NEXT_PUBLIC_API_URL` pointed at prod without `/v1`; `api/.env` `DATABASE_URL` pointed at a Postgres that isn't running. Both files gitignored, so invisible to git |
| Deployed Google login opens consent, then dumps you back on the landing page | web + api | code trace, 2.0 (b) | **Cookie set on the API host, read from the web host.** `SetTokenCookies` scopes `access_token` to `surus-1.vercel.app`; `getServerSession` (`web/lib/auth/server.ts:8`) reads the web origin's jar and finds nothing, so `(auth)/layout.tsx` redirects to `/`. Not fixable with `Domain=` — `vercel.app` is on the Public Suffix List. See decision 2. **Both halves of the fix (2.7a, 2.7b) now land in code — schema change not yet applied to any database, and nothing has been exercised end-to-end (2.1 still blocks booting the API at all). Do not mark this row fixed until a real login has been observed to work.** |
| Every session dies 15 min after login, in every environment | api | code inspection, `service.go:218` | **`RefreshSession` is a stub** — hardcoded `unauthenticated` return, no lookup implemented. `access_token` `MaxAge: 900` and nothing can renew it. **Fixed in 2.8, still unverified** — proving it needs a deliberate wait past the 900s TTL, which the 2026-08-13 pass did not do. |
| Every `/v1/*` call 404s | web | code trace, 2.0 (a) | `NEXT_PUBLIC_API_URL` has no `/v1` suffix; fetch layers concatenate `` `${API_URL}${path}` `` with no normalization. Also check the Vercel dashboard value |
| API exits on boot locally | local | `docker ps` fails; `main.go:42-45` | `DATABASE_URL` → `localhost:5432`, no Postgres running. `pool.Ping` fails, `os.Exit(1)` |
| Every magic link rejected as "Invalid or expired token" | api | code trace, 2.0 (d) | **Hash mismatch.** Token stored as a salted bcrypt digest; `VerifyMagicLink` looks it up with the raw token against `WHERE token_hash = $1`. Never matches. Same bug class as the refresh stub. **Fixed in 2.10, still unverified** — the flow cannot be walked at all until an email sender exists. |
| Magic-link emails never sent | api | `handler.go:59-82` | `MagicLinkRequest` discards the raw token into `_` and returns 204; the Postmark call is a comment. `POSTMARK_*` env vars are configured but unread. **Still open** — deliberately out of scope for the 2.10 hash fix, needs a product/config decision first, see Completed note |
| OAuth `state` never verified (CSRF) | api | `handler.go:165` | `state` query param read into `_` — no check at all. **Fixed in 2.11, confirmed** — a real Google login on 2026-08-13 passed the new check rather than failing closed. The cookie's `SameSite=Lax`/no-`Secure` was *not* an additional bug; see the correction in 2.0 |
| `POST /v1/auth/google/callback` accepts `state` unverified and is dead code | api | `routes.go:13`, `handler.go:68`; grep of `web/` finds no caller | 2.11 only covered `GoogleCallbackRedirect`. This JSON twin is **mounted and live in production**, reads `state` from the body, never compares it to the cookie, and issues session cookies. Nothing in `web/` calls it. Its partner `GoogleStart` (`handler.go:22`) isn't even routed. Recommend deleting both plus the route line rather than hardening dead code — see 2.12 |
| ~~Course list empty/failing after login~~ **RESOLVED** | web + api | verified 2026-08-13 | Was downstream of the env drift, exactly as presumed. Course list renders after login |
| ~~Course creation~~ **RESOLVED** | web + api | verified 2026-08-13 | Same root cause. Creation, listing, and search all work |
| ~~Both hardcoded "Sign in with Google" links 404~~ — **NOT A BUG, claim retracted** | web | 2.7b claimed this; disproved by `routes.go:11` | `/google/start` is mounted **twice**: at root via `main.go:95` *and* under `/v1/auth` via `routes.go:11`, both hitting `GoogleStartRedirect`. So `/v1/auth/google/start` resolves fine and the old links worked — consistent with the user's report that OAuth does open in production. The 2.7b rewrite is still an improvement (handles a trailing slash, supplies a default) but fixed no 404. **Corollary: production's `NEXT_PUBLIC_API_URL` almost certainly already ends in `/v1`**, since the old link only worked if it did |
| ~~Logout leaves a zombie session on the web origin~~ — **fixed in 2.7c** | web | code trace, 2.7c | `useAuth`'s `logout()` only called `POST /auth/logout` via `apiFetch`, clearing only the **API**-origin jar. Fixed by routing sign-out through a new `web/app/auth/logout/route.ts` Route Handler instead. |

The rows marked "code inspection" / "code trace" were found by reading the code, **not** by reproducing the failure. They are real bugs regardless. 2.3 still has to confirm which of them actually produce the symptoms you see, and the last two rows stay unconfirmed until the environment is fixed.

---

## Completed

**1.1 MCP servers (2026-08-13).** Created `.mcp.json` with only the three main-chat servers (`context7`, `filesystem`, `sentry`). Per-agent scoping uses the subagent `mcpServers:` frontmatter field, which accepts full inline server definitions — so agent-only servers are never registered in the main session rather than being registered-then-hidden. Each agent additionally carries a `tools:` allowlist using `mcp__<server>__*` patterns.

**1.2 Custom agents (2026-08-13).** Wrote `.claude/agents/{backend,frontend,debug,planner,lookup}.md`. Directory ownership is enforced in the prompts (`backend` → `api/`, `frontend` → `web/`), `debug` has no write tools, and `planner` may write only `PLAN.md`. Models are pinned per agent rather than inherited so a long main-thread session on Opus doesn't drag every subagent onto it: the three working agents run Sonnet 5, `planner` runs Opus 4.8, and `lookup` runs Haiku 4.5. The planned `deploy` agent was dropped once 1.4 established that deployment is Vercel dashboard configuration plus a CI gate, with nothing for an agent to drive. All five frontmatter blocks were parsed with js-yaml to confirm they load.

**1.3 Scoped rules (2026-08-13).** Split the root `CLAUDE.md` into `api/CLAUDE.md` and `web/CLAUDE.md` and trimmed the root to cross-cutting material only — handoff protocol, agent roster, commands, deploy topology, branch facts. Detail was moved, not copied, so there is one home per fact. The cross-origin cookie/CORS rules are stated in both scoped files because both sides have to honor them, but from each side's perspective.

**1.4 Vercel deploy, repo side (2026-08-13).** Added `web/vercel.json` and `api/vercel.json` with an `ignoreCommand` that diffs the project's own root directory against `VERCEL_GIT_PREVIOUS_SHA`, so neither project rebuilds for the other's changes. Added `.github/workflows/ci.yml` as a build/lint gate only — deployment stays on Vercel's native Git integration, since Deploy Hooks would duplicate it and lose the commit association.

While implementing the cross-origin constraints, four real bugs surfaced by inspection and were fixed in `api/`:

- `middleware.CORS` took a single origin and set it unconditionally. It now takes a list, echoes the request's `Origin` only when allowlisted, and always sends `Vary: Origin`. It also supports one `*` wildcard per entry; that was written for per-branch preview hostnames, and since previews are build-only and never sign anyone in, it should stay unused — a bare `https://*.vercel.app` would let any site on that domain make credentialed requests. The warning lives in the code comment and `.env.example`.
- `access_token` and `refresh_token` were `SameSite=Lax`, which browsers withhold on cross-site requests. `Secure` and `SameSite` are now chosen together in `cookieAttrs()` — `None`+`Secure` when deployed, `Lax` without `Secure` for plain-HTTP localhost, since browsers reject `None` without `Secure`.
- `refresh_token` was scoped `Path=/auth/refresh` while the route is mounted at `/v1/auth/refresh`, so the browser never sent it and refresh could not have worked in any environment. Now a named constant.
- `ClearTokenCookies` omitted `Secure`/`SameSite`, so its deletion cookie didn't match the original and would append a second cookie rather than overwrite it.

`go build ./...` and `go vet ./...` pass. `pnpm build` passes. None of this is verified against a running browser yet — that is 2.2.

**2.8 Implement `RefreshSession` (2026-08-13).** `GetRefreshTokenByHash` (`api/db/queries/auth.sql`) already existed and already does a direct indexed `WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()` lookup — no new query or migration needed, so `sqlc generate` was not run.

The real blocker was that `GenerateRefreshToken` (`api/internal/auth/jwt.go`) hashed the raw token with **bcrypt**, a salted, non-deterministic hash — you cannot look one up by equality; the stub's own comment described iterating every non-revoked row and bcrypt-comparing each, which is what made the query unusable and was presumably why the stub was never finished. Switched `GenerateRefreshToken` to a deterministic **SHA-256** hex digest instead (`HashRefreshToken`, new exported helper). This is safe because the raw token is already 256 bits from `crypto/rand` — a salt defends against low-entropy secrets like passwords, which this isn't — so a plain digest supports a direct indexed lookup with no security loss. Removed the now-unused bcrypt-based `VerifyRefreshToken`; nothing called it. `GenerateMagicLinkToken` still uses bcrypt and was left untouched (out of scope) — note `VerifyMagicLink` looks broken the same way (hashes the raw token with bcrypt on write, then queries `token_hash = $1` with the raw token on read, which can never match); did not touch it since it's outside 2.8.

`RefreshSession` (`api/internal/auth/service.go`) now: hashes the presented raw token, looks it up via `GetRefreshTokenByHash` (a `pgx.ErrNoRows` — which the query's own WHERE clause maps expired/revoked/missing all onto — becomes `unauthenticated`, any other DB error becomes `internal_error` per the error contract), loads the user via `GetUserByID`, issues a new pair via the existing `issueTokens`, then **rotates**: calls `RevokeRefreshToken` on the row that was just used so it can't be replayed. Returns `(*db.User, *TokenPair, error)` unchanged, so `handler.go`'s `Refresh` needed no edits.

`go build ./...` and `go vet ./...` pass from `api/`. **Not exercised at runtime** — `api/.env` `DATABASE_URL` points at a dead local Postgres (2.1), so the server cannot boot locally, and I did not touch `.env` or attempt to point it at Neon. This needs an end-to-end check once 2.1 is done: log in, wait past `access_token`'s 900s `MaxAge`, hit `/v1/auth/refresh`, confirm a new `access_token` is set and the old `refresh_token` is rejected on a second use (rotation).

One side effect worth flagging: any `refresh_tokens` rows already in the Neon prod branch were written with the old bcrypt hash format and will no longer match `GetRefreshTokenByHash` after this change, since the stored hash format changed. This is not a regression — those rows were already fully unusable (the stub always failed before), so nothing that worked before stops working — but it does mean stale bcrypt rows will sit in the table until their `expires_at` (30 days) passes. No cleanup was performed as part of this task since it is not a security or correctness issue, just table hygiene.

**2.10 (hash half) + 2.11 (2026-08-13).** `api/db/queries/auth.sql`'s `GetMagicLinkTokenByHash` was already correct — same shape as the refresh-token query, `WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()` — so no `.sql` edit and no `sqlc generate`. The only defect was on the write side: `GenerateMagicLinkToken` (`api/internal/auth/jwt.go`) hashed with salted bcrypt, so the digest could never equal itself on lookup. Switched it to the same deterministic SHA-256 scheme as `GenerateRefreshToken` (new `HashMagicLinkToken` helper, mirroring `HashRefreshToken`), and `VerifyMagicLink` (`api/internal/auth/service.go`) now hashes the presented raw token before calling `GetMagicLinkTokenByHash` instead of passing the raw token straight through. `bcrypt` import removed from `jwt.go`; confirmed with a repo-wide grep that nothing else in `api/` imports `golang.org/x/crypto/bcrypt`, so no other flow depended on it. Also changed `VerifyMagicLink`'s failure code from `validation_error` to `unauthenticated` per the error contract — an invalid/expired credential isn't a request-shape problem. Confirmed `MarkMagicLinkTokenUsed` already ran before `issueTokens` in the existing code order, so redemption is already single-use; left that sequencing untouched.

**Did not wire up Postmark.** `MagicLinkRequest` (`handler.go`) still discards the raw token and returns 204 without sending anything — that's the other half of 2.10 and is a pending product/config decision (Postmark account, sender domain verification, from-address), not a bug fix, per the task scope. Wiring it up would need: a small `internal/mail` (or similar) package wrapping the Postmark HTTP API using `POSTMARK_API_KEY`/`POSTMARK_FROM_EMAIL` (already in `.env.example` but unread), a template with the magic-link URL (`{magicLinkBaseURL}?token={raw}`), and a decision on whether to fail the request or swallow-and-log if the send errors (right now the token is generated and stored regardless, so probably swallow-and-log to avoid leaking whether an email exists via response timing/status — worth an explicit decision, not just an implementation detail).

2.11: `GoogleCallbackRedirect` (`api/internal/auth/handler.go`) now reads `state` from the query string, compares it against the `oauth_state` cookie with `crypto/subtle.ConstantTimeCompare`, and redirects to `h.frontendURL+"?error=auth_failed"` on a missing cookie or mismatch, before ever calling `ExchangeGoogleCode`. Added `Service.CookieAttrs()` (thin exported wrapper around the existing unexported `cookieAttrs()`) so the handler doesn't duplicate the Secure/SameSite decision, and two handler-local helpers, `setOAuthStateCookie`/`clearOAuthStateCookie`, that both `GoogleStart` and `GoogleStartRedirect` now call instead of hand-rolling a hardcoded `SameSite=Lax`, no-`Secure` cookie. **Corrected:** this entry originally claimed the shared attributes also fixed a dropped cookie, on the grounds that `SameSite=Lax` without `Secure` "gets dropped on the cross-site return from Google." That reasoning is wrong — Lax cookies survive cross-site top-level GET navigations, which is what the redirect is, and no-`Secure` cookies are still sent over HTTPS. The old `oauth_state` cookie arrived fine; the only defect was the missing check. See the correction in 2.0. Sharing the attributes is still correct, but for a different reason: the JSON pair below reaches the callback via `fetch`, where Lax genuinely would withhold it. `GoogleCallback`/`GoogleStart` (the JSON, non-redirect pair) share the same `setOAuthStateCookie` helper now too, so their cookie has consistent attributes; `GoogleCallback` itself still reads `state` from the JSON body rather than checking it against the cookie — it never did, and adding that check was outside what 2.11 described (only `GoogleCallbackRedirect` was named), so it's left as-is. Worth a follow-up look if that JSON pair is actually used anywhere, since it currently accepts `state` without verifying it at all.

`go build ./...` and `go vet ./...` pass from `api/`. **Not exercised at runtime** — same reason as 2.8, the API cannot boot locally against a dead Postgres. Needs an end-to-end check once 2.1 is done: request a magic link, confirm `token_hash` in `magic_link_tokens` is a 64-char hex string (not a `$2...` bcrypt string), verify it, confirm a second verify attempt with the same token fails; and separately, run the Google login redirect flow and confirm a request to `/auth/google/callback` with a forged/missing `state` bounces to `?error=auth_failed` instead of exchanging the code.

**Any existing rows in `magic_link_tokens` on the Neon branch are now unmatchable**, same as the refresh-token note above — they were written with the old bcrypt `token_hash` format and will never match the new SHA-256 lookup. Not a regression (those rows were already fully broken, since the raw-vs-bcrypt mismatch predates this fix), and they self-expire in 15 minutes per `RequestMagicLink`'s `ExpiresAt`, so there's nothing to clean up.

**1.5 Clear `no-explicit-any` lint errors, make lint gate blocking (2026-08-13).** The task's premise was slightly off: `pnpm lint` had 19 errors total, but only 15 were `@typescript-eslint/no-explicit-any` — the other 4 were `react-hooks/set-state-in-effect` (React's newer "don't setState synchronously inside an effect" rule), pre-existing and unrelated to `any`. Fixed both sets, since a blocking gate needs `pnpm lint` to exit 0 regardless of which rule is firing.

`no-explicit-any` fixes, all real types (no `unknown` swaps, no disables):
- `lib/types.ts` — `TiptapDoc.content` was `any[]`. Added `TiptapNode`/`TiptapMark` interfaces mirroring Tiptap's own `JSONContent` shape (`type`, `attrs?`, `content?`, `marks?`, `text?`) rather than importing `@tiptap/core` directly, since it's not a direct dependency under pnpm's strict `node_modules` and isn't guaranteed to resolve.
- `app/(public)/page.tsx`, `app/(public)/users/[userId]/page.tsx` — `courses: any[]` / `course: any` replaced with `Course` from `lib/types` (the server fetch layer already returned it typed; the `any` was only in the local variable annotation).
- `components/course/LessonTree.tsx` — the recursive tree node was typed inline as `Lesson & { children: Lesson[] }` at the leaf but `any` on the recursive `.map` callback. Added a named recursive type `TreeLesson = Lesson & { children: TreeLesson[] }` and used it throughout; this also fixed a latent type gap the `any` had been masking (`.children` was actually `Lesson[]`, not `TreeLesson[]`, until the type was made properly recursive).
- `components/lesson/VideoLesson.tsx` — `ComponentType<any>` cast on the dynamically-imported `react-player` default export was unnecessary; removing it exposed that the installed `react-player@3.4.0`'s prop is `src`, not `url` (the component was passing a prop the library doesn't have — a real, previously-hidden bug, now fixed).
- `components/shared/ReportButton.tsx` — `category as any` (x2) replaced by typing the `category` state as `ReportInput["category"]` from `lib/types`, with a single narrowing cast where the radio input's `e.target.value` (a bare `string`) is assigned to it.
- `app/(auth)/courses/new/page.tsx`, `app/(auth)/courses/[courseId]/edit/page.tsx` — four `any`s across the two files, all on lesson/course update payloads (`updates: any`, `input: any`, one `onValueChange`/`v: any`). Typed against `Partial<CourseInput>` / `LessonInput` (already defined in `lib/types.ts` and already used by `lib/api/courses.ts` / `lib/api/lessons.ts`, just not propagated into these two client components). Tightening these surfaced two more latent bugs `any` had hidden: the `visibility` local state was `useState<string>`, which doesn't satisfy `CourseInput`'s literal union, and spreading `{...lesson.video, curator_notes: doc}` doesn't type-check when `lesson.video` is optional (TS makes every spread property optional) and separately has a `number | null` vs. `number | undefined` mismatch between `VideoDetail` (API response shape) and `LessonInput.video` (API request shape) for `start_seconds`/`end_seconds`. Fixed both: `visibility` state now uses the literal union, and both files build an explicit `baseVideo` object that normalizes `null` to `undefined` before spreading.

`react-hooks/set-state-in-effect` fixes (4, not part of the original ask but required for a clean, blocking `pnpm lint`): both course-editor pages had a per-lesson `LessonEditor` subcomponent syncing `title`/`videoUrl` state from the `lesson` prop via `useEffect(() => {...}, [lesson.id])`. Replaced with React's documented alternative — add `key={lesson.id}` (`selectedLesson.id`) at the call site and initialize state directly from the prop, so switching lessons remounts the editor instead of effect-syncing it. The course-edit page additionally had two effects syncing `title`/`description`/`visibility` from `courseData` and `lessons` from `lessonsData` (react-query results) — these can't use the `key` trick since the whole page isn't keyed on course identity, so they use React's other documented pattern, "adjust state during render": a `prevX` state variable compared against the incoming value, with the `setX` calls made directly in the render body (guarded by the comparison) rather than in a `useEffect`. Both patterns are from https://react.dev/learn/you-might-not-need-an-effect.

Also removed now-dead `useEffect` imports in both page files once their only two effects were gone.

Verification: `pnpm lint` — 0 errors, 21 pre-existing warnings (unused imports/vars, `<img>` vs `next/image`), exit 0. `pnpm build` — passes (`NEXT_PUBLIC_API_URL` set inline for the build only, `.env.local` untouched). Removed `continue-on-error: true` from the `pnpm lint` step in `.github/workflows/ci.yml`, so the web CI job now fails on any lint error. **Did not exercise the running app** — per the task's note, the API points at a dead database and `NEXT_PUBLIC_API_URL` is misconfigured (2.0/2.1/2.9), so none of this was checked against a live browser session; the course editor and lesson-tree/video-lesson rendering paths touched here are exactly the ones flagged `todo` in 2.4–2.6 and should be re-verified once those are unblocked.

No API-side gaps found — every `any` had a real type available on the `web/` side already (either in `lib/types.ts` or derivable from it) without needing a shape change in `api/`.

**2.7b Web callback route (2026-08-13).** Added `web/app/auth/callback/route.ts`, a `GET` Route Handler outside the `(auth)` route group (confirmed by inspection: `(auth)/layout.tsx`'s session guard only wraps files under `app/(auth)/`; `app/auth/callback/` is a sibling literal segment, and `pnpm build`'s route table lists `/auth/callback` on its own with no group prefix — so the guard genuinely doesn't apply here). It reads `?code=` from the query string, `POST`s `{"code"}` to `${NEXT_PUBLIC_API_URL}/auth/handoff` **from the server** (plain server-side `fetch`, never exposed to the browser), and on a 200 with both `access_token` and `refresh_token` present, sets both as cookies on the web origin via `cookies()` and redirects to `/dashboard`. Any failure — missing `code`, non-2xx, a body missing either token, or a thrown network error — redirects to `/?error=auth_failed` and sets nothing; the API's own error message is never forwarded into the URL, only the fixed `auth_failed` string.

Cookie attributes mirror `api/internal/auth/service.go`'s `cookieAttrs()`/`SetTokenCookies` exactly (read, not guessed): `httpOnly: true`; `secure`/`sameSite` chosen together — `Secure` + `SameSite=None` when `NEXT_PUBLIC_API_URL`'s resolved protocol is `https:`, else no `Secure` + `SameSite=Lax` for plain-HTTP localhost, matching the API's `appEnv === "development"` branch but driven off the URL's protocol instead of an env var, since this is the web project and doesn't read `APP_ENV`; `access_token` at `path: "/"`, `maxAge: 900`; `refresh_token` at `path: "/v1/auth/refresh"`, `maxAge: 2592000`. The `refresh_token` path is mirrored for consistency with the API's own cookie even though nothing on the web origin currently reads a cookie at that path — noted in a code comment so a future reader doesn't assume it's load-bearing on the web side.

**Contract assumed, not yet verified against a live endpoint:** `POST {NEXT_PUBLIC_API_URL}/auth/handoff` → `{"access_token", "refresh_token"}` on 200, `{"error": {...}}` on 4xx, per the task brief. Grepped `api/internal/auth` for `handoff`/`Handoff` and found nothing — the backend half of 2.7 (the handoff endpoint itself) was not yet implemented as of this pass, only schema/query-level references in `db/queries/auth.sql` and the migration. This route cannot be exercised end-to-end until that lands.

**Update (2026-08-13, backend, 2.7a done): the assumed contract above is now implemented exactly as written** — see the `2.7a` Completed entry below for the exact request/response shape and error codes. Still not exercised end-to-end (see that entry for why).

**Found and fixed two more broken sign-in links while doing the 2.7 task's step 4 check.** `components/shared/Nav.tsx` and `components/shared/AuthPrompt.tsx` both built the Google sign-in URL by stripping `/v1` off `NEXT_PUBLIC_API_URL` and then **re-appending** `/v1/auth/google/start` — e.g. `` `${apiBase}/v1/auth/google/start` `` where `apiBase` was already `/v1`-stripped. **The 404 claim in this paragraph is retracted — it was wrong.** `/google/start` is mounted **twice**: at root (`api/cmd/server/main.go:95`) *and* under `/v1/auth` (`api/internal/auth/routes.go:11`), both pointing at `GoogleStartRedirect`. `/v1/auth/google/start` therefore resolves, the old links worked, and this matches the user's report that production OAuth does open the consent screen. The rewrite is still worth keeping — it strips a trailing slash correctly and supplies a default when the env var is unset — but it fixed no outage. Corrected rather than deleted so nobody re-derives the same wrong conclusion. Corollary worth carrying forward: production's `NEXT_PUBLIC_API_URL` almost certainly already ends in `/v1`, because the old link could only have worked if it did — which makes the cookie split (2.0 (b)) the sole remaining production blocker, not a second 404.

**Found, not fixed: logout does not clear the web-origin cookie jar.** `hooks/useAuth.ts`'s `logout()` calls `lib/api/auth.ts`'s `logout()`, which is `apiFetch("/auth/logout", ...)` — `credentials: "include"` against the **API** origin, so it only triggers `ClearTokenCookies` on the API's own jar. The web-origin `access_token`/`refresh_token` cookies set by this task's new callback route are a separate jar and are never touched by that call, so a "logged out" user still carries a valid web-origin `access_token` for up to 900s (and a `refresh_token` for up to 30 days, though nothing web-side currently uses it to refresh). Per the task brief, did not fix this myself — added subtask 2.7c rather than guessing at the right shape (e.g., whether it should be a new web-side route that clears its own cookies, or whether `useAuth.logout` should also call it). No API change is needed for a fix; the gap is entirely on the web side, but the task instructions asked me to flag rather than resolve it within 2.7's scope.

Verification: `pnpm lint` — 0 errors, 21 pre-existing warnings, same count as before this change (no new warnings introduced). `pnpm build` — passes; route table shows `ƒ /auth/callback` as an independent dynamic route, not nested under the `(auth)` group. **Did not exercise end-to-end.** `web/.env.local` still has `NEXT_PUBLIC_API_URL=https://surus-1.vercel.app/` (no `/v1`, points at prod — unresolved 2.9) and `api/.env`'s `DATABASE_URL` still points at a dead local Postgres (unresolved 2.1), so the local API cannot boot; and even if it could, `/auth/handoff` isn't implemented yet (see above). This needs a real run once 2.1, 2.9, and the backend half of 2.7 all land: click "Sign in", complete Google consent, confirm the browser lands on `/dashboard` with both cookie jars populated, and separately confirm `?code=` reuse or an expired/invalid code redirects to `/?error=auth_failed` without leaking the API's error text.

**2.7c Web-side logout clears both cookie jars (2026-08-13).** Added `web/app/auth/logout/route.ts`, a `GET` Route Handler (not `(auth)`-grouped, sibling literal segment like `auth/callback`) that: reads the web-origin `access_token` cookie directly via `cookies()`; if present, best-effort `POST`s `${NEXT_PUBLIC_API_URL}/auth/logout` with a manually-built `Cookie: access_token=...` header (same forwarding pattern as `serverFetch` in `lib/api/server/courses.ts` — read before writing, not invented) so the API can revoke the user's refresh tokens and handoff codes server-side; wrapped in try/catch that swallows any failure; then unconditionally deletes both `access_token` (`path: "/"`) and `refresh_token` (`path: "/v1/auth/refresh"`) from the web-origin jar — same paths `app/auth/callback/route.ts` used when setting them, since a mismatched `path` on deletion just writes a second cookie instead of overwriting (the exact bug 1.4 fixed on the API side); then redirects to `/`. The API call and the cookie deletion are sequenced so deletion always runs, network failure or not — a user who clicks sign out ends up logged out locally regardless of API reachability.

Changed `useAuth`'s `logout` (`web/hooks/useAuth.ts`) from an `apiFetch`-based async function into a synchronous `window.location.href = "/auth/logout"` — this has to be a top-level navigation, not a `fetch`, because a fetch response's `Set-Cookie` can't reach the browser's own jar for this origin the way a navigated response can. Removed the now-dead `logout()` export from `lib/api/auth.ts` (only caller was `useAuth`) and the `queryClient` cache-clearing logic in the old `logout` (moot — the navigation causes a full page load, which resets all client state including the TanStack Query cache, so nothing needs to invalidate it manually).

Updated both call sites: `components/shared/Nav.tsx`'s `handleLogout` now just calls `logout()` (no more `await` + `router.push("/")`, since the route handler redirects); `app/(auth)/settings/page.tsx`'s account-deletion flow (`deleteMutation.onSuccess`) does the same, dropping the now-unused `useRouter` import.

`pnpm lint` — 0 errors, 21 pre-existing warnings, same set as before this change. `pnpm build` — passes; route table lists `ƒ /auth/logout` as an independent dynamic route alongside `ƒ /auth/callback`, confirming the `(auth)` layout's session guard doesn't wrap it. **Did not exercise end-to-end** — same blocker as every other 2.7 sub-entry: `web/.env.local`'s `NEXT_PUBLIC_API_URL` still has no `/v1` and points at prod (2.9 open), and `api/.env`'s `DATABASE_URL` still points at a dead local Postgres (2.1 open), so the local API cannot boot. Needs a real run once those land: sign in, sign out, confirm landing on `/`, confirm `/dashboard` afterwards redirects rather than rendering, and confirm the API actually received the `/auth/logout` call (revoked refresh token/handoff codes) rather than just the local cookie clear succeeding on its own.

No `api/` changes needed or made — the gap was entirely on the web side as 2.7b's Completed note predicted.

**2.7a API handoff endpoint (2026-08-13, backend).** Implements the API side of decision 2 / option B, matching what 2.7b already assumed.

New table `handoff_codes` (`api/db/migrations/001_initial_schema.sql`): `id, user_id, code_hash UNIQUE, expires_at, used_at, created_at`, plus a partial index on `user_id WHERE used_at IS NULL`. Same shape as `magic_link_tokens`/`refresh_tokens` — hash stored, never the raw code. **This is a schema change that must be applied by hand to whichever database is in use; there is no migration runner. I did not apply it to the Neon production branch or any other branch — it only exists in the migration file and in generated Go code until someone runs it against a real database.** Three new queries in `api/db/queries/auth.sql` (`CreateHandoffCode`, `RedeemHandoffCode`, `RevokeAllUserHandoffCodes`); `sqlc generate` was run from the repo root (via `go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate`, since no `sqlc` binary was on PATH) and only touched generated files plus line-ending normalization on unrelated `.sql.go` files (confirmed via `git diff --stat` — 4 files with real content changes, the rest CRLF/LF only).

`RedeemHandoffCode` is `UPDATE handoff_codes SET used_at = now() WHERE code_hash = $1 AND used_at IS NULL AND expires_at > now() RETURNING ...` — a single conditional UPDATE, not a SELECT-then-UPDATE, so a concurrent double-submit of the same code can only have one winner; Postgres's row-level locking on the UPDATE is the atomicity guarantee.

`api/internal/auth/jwt.go`: added `GenerateHandoffCode`/`HashHandoffCode` (32 bytes from `crypto/rand`, SHA-256 hex digest, same deterministic pattern as `GenerateRefreshToken`/`GenerateMagicLinkToken` — no bcrypt, so no repeat of the 2.0(d)/2.8 hash-mismatch bug class) and `HandoffCodeTTL = 60 * time.Second`.

`api/internal/auth/service.go`: added `CreateHandoffCode(ctx, userID) (string, error)` (mints + stores, returns the raw code) and `RedeemHandoffCode(ctx, rawCode) (*db.User, *TokenPair, error)` (hashes, atomically claims the row, `pgx.ErrNoRows` → `unauthenticated` "Invalid or expired handoff code", any other DB error → `internal_error`, then issues a **fresh** token pair via the existing `issueTokens` — the handoff code itself is never a session credential, only a one-time key to mint one). `Logout` now also calls the new `RevokeAllUserHandoffCodes`, so a stolen-but-unredeemed code can't resurrect a session after the user logs out.

`api/internal/auth/handler.go`: `GoogleCallbackRedirect` keeps its existing `SetTokenCookies` call unchanged (still populates the API-origin jar, since the callback is a top-level navigation to the API origin), then additionally calls `CreateHandoffCode` and redirects to `h.frontendURL + "/auth/callback?code=" + url.QueryEscape(handoffCode)` instead of `/dashboard`. On a `CreateHandoffCode` error it falls back to `?error=auth_failed` rather than leaving the user on a dead redirect. New `Handoff` handler for `POST /v1/auth/handoff` — see contract below. `api/internal/auth/routes.go` adds `r.Post("/handoff", h.Handoff)` inside `authHandler.Routes`, which is mounted under `/v1/auth` in `main.go` behind the existing 10-req rate limiter (`r.Route("/auth", func(r chi.Router) { r.Use(middleware.RateLimit(authRL)); r.Mount("/", authHandler.Routes(authCfg)) })`) — confirmed by reading `main.go`, not assumed; the new route needed no separate wiring to inherit the limit.

**Exact contract for `POST /v1/auth/handoff`** (frontend: this is what `app/auth/callback/route.ts` should call, and it already does, unmodified):

- Request: `Content-Type: application/json`, body `{"code": "<the code query param from the /auth/callback redirect, url-decoded>"}`. No auth cookie required or read — the code itself is the credential.
- Success: `200`, body `{"access_token": "<raw JWT>", "refresh_token": "<raw opaque token>"}`. **No cookies are set on this response** — the caller (the Next server) is expected to set its own cookies from the body, since the caller is not a browser and `Set-Cookie` on a server-to-server response wouldn't reach the actual browser anyway.
- Failure: standard error envelope `{"error": {"code": "...", "message": "..."}}` — `400 validation_error` for a missing/malformed body or missing `code`, `401 unauthenticated` for a code that's missing, already used, or past its 60s TTL, `500 internal_error` for anything else (DB failure, etc.). Codes and statuses come from `middleware.ServiceError`/`statusMap`, not ad hoc.
- The code is single-use: a second call with the same code, concurrent or sequential, gets `401 unauthenticated` on every call after the first success.
- Rate limiting: inherits the existing `/v1/auth/*` 10-req limiter, no separate configuration.

**Verification: `go build ./...` and `go vet ./...` both pass from `api/`.** `go run ./cmd/server` was attempted and **fails to boot** — `api/.env`'s `DATABASE_URL` still points at `localhost:5432` with nothing listening there (2.1 is still `todo`), so `pool.Ping` fails and the process exits before serving a single request. **Did not exercise `/health` or `/v1/auth/handoff` at runtime; nothing in this entry has been checked against a live server.** This needs a real end-to-end pass once 2.1 lands: run the full Google OAuth flow, confirm the API-origin cookies are set, confirm the redirect carries a `?code=` and never a raw JWT, `POST` that code to `/v1/auth/handoff` and confirm the token pair comes back, confirm a second `POST` with the same code returns `401`, and confirm a code left unredeemed for 60s also returns `401`.

One thing worth flagging for whoever runs the schema change: `handoff_codes.code_hash` is `UNIQUE`, same as `magic_link_tokens.token_hash` — a SHA-256 collision on 256 bits of `crypto/rand` input is not a practical concern, this just mirrors the existing pattern rather than introducing a new one.

---

## Open decisions

| # | Decision | Blocking |
| - | -------- | -------- |
| 1 | Linkrot job: Vercel Cron endpoint vs. drop the feature | 1.4 |
| 2 | How to make the session cookie readable by the web app | 2.7, and all of Task 2 in production |

### Decision 2 — cross-domain session cookie

The API sets `access_token` on its own host; the web app reads cookies from its host; on
`*.vercel.app` those can never be the same jar (Public Suffix List blocks a shared `Domain`).
Three ways out, in recommended order:

**A. Custom domain.** `surus.com` → web project, `api.surus.com` → api project. The API sets
`Domain=.surus.com`, both origins share the cookie, and `SameSite=Lax` is enough — the cross-site
problem stops existing rather than being worked around. Also gives Google OAuth one permanent
redirect URL, which is the same constraint that makes previews unusable for testing (1.4). Cost: a domain purchase
and DNS setup. Changes needed in code are small (a `COOKIE_DOMAIN` env var read in `cookieAttrs`).

**B. Web-side callback handoff.** `GoogleCallbackRedirect` redirects to a Next Route Handler on the
web origin (`/auth/callback?token=…`) which sets the cookie itself. No domain needed. Downside: a
JWT travels in a URL, so it lands in browser history, server access logs, and `Referer` headers.
Mitigable with a short-lived single-use exchange code, which is more code than A.

**C. Next rewrite proxy.** Web rewrites `/api/*` to the API origin so the browser only ever sees one
origin. Contradicts `web/CLAUDE.md` ("never a same-origin `/api/*` route"), routes all API traffic
through Next functions, and adds a hop to every request.

Recommendation was **A**. **Decided 2026-08-13: ship B now as a bridge, move to A when the domain
is purchased.** A is still the destination — B is explicitly temporary and carries debt (below).

#### Correction to option B as originally written

The sketch above ("a Next Route Handler sets the cookie itself") is **incomplete and would break
client-side fetches.** There are two fetch layers with two different cookie needs:

| Layer | Runs where | Reads which jar |
| ----- | ---------- | --------------- |
| `serverFetch`, `getServerSession` (`web/lib/api/server/*`, `web/lib/auth/server.ts`) | Next server | **web** origin — via `cookies()` |
| `apiFetch` (`web/lib/api/client.ts`) | browser, `credentials: "include"` | **API** origin |

Setting the cookie only on the web origin fixes the first and breaks the second. So B means the
same JWT lives in **both** jars:

1. `GoogleCallbackRedirect` keeps its existing `SetTokenCookies` call — that still works, because
   the callback is a top-level navigation to the API origin.
2. It *additionally* redirects to a new web-side route with a short-lived single-use handoff code.
3. That web route exchanges the code with the API and sets its own copy of `access_token` on the
   web origin.

**Debt this creates, to be paid off by A:** two cookies holding the same token, two expiries to keep
in sync, and logout must clear both or a zombie session survives on one origin. The handoff code
must be single-use and short-lived (≤60s) — do **not** put the raw JWT in the redirect URL, where it
would land in browser history, access logs, and `Referer` headers.

Resolved:

- **Preview environments — kept, but build-only (2026-08-13, corrected same day).** Vercel previews are enabled and run on PRs; an earlier revision wrongly recorded them as "dropped." They are build verification, not a test environment: Google OAuth rejects unregistered redirect URLs and preview hostnames rotate, so a preview login can never complete. Functional testing is local against a Neon branch. Accepted tradeoff: no staging gate, merging to `main` is live.
- `deploy` agent MCP choice — agent dropped.
- MCP credential storage — all OAuth or localhost, nothing to commit.

## Reference

- [Run any Dockerfile on Vercel](https://vercel.com/blog/dockerfile-on-vercel)
- [Bring your Dockerfile to Vercel Functions](https://vercel.com/changelog/bring-your-dockerfile-to-vercel-functions)
- [Vercel Deploy Hooks](https://vercel.com/docs/deploy-hooks)
- [Deploying Git repositories with Vercel](https://vercel.com/docs/git)
- [Claude Code subagents — frontmatter reference](https://code.claude.com/docs/en/sub-agents)
