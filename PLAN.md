# PLAN

Task board. Durable facts about how the code works live in the `CLAUDE.md` files; current work state lives here. Handoff protocol: `CLAUDE.md` → "Agent handoff protocol".

Last updated: 2026-08-14

---

## Status

The MVP works locally: Google sign-in, course creation, course listing, and search all verified in a browser against a Neon dev branch (2026-08-13).

**Production is not yet working.** Sign-in half-completes and course creation fails. Cause identified — see "In flight".

Four PRs are on `main`: dev tooling and docs (#5), strict types plus a blocking lint gate (#6), auth session fixes (#8), and the cross-origin session bridge (#9).

---

## In flight

| # | Task | Status |
| - | ---- | ------ |
| A | Vercel redeploy fix + origin normalization | PR open, awaiting merge |

### A — why production is broken

Two bugs stacked, both in config this repo owns.

**1. `NEXT_PUBLIC_APP_URL` on the api project had a trailing slash.** That value is both concatenated into redirect URLs and used to seed the CORS allowlist, which `middleware.originAllowed` matches *exactly* against the browser's `Origin` header. An `Origin` never carries a trailing slash, so every cross-origin request was rejected.

That splits the symptoms exactly as observed: the server-rendered auth guard reads the web-origin cookie and passes, so you reach the dashboard — then every call on the page goes through `apiFetch` in the browser, gets blocked, and the page looks logged out. Course creation fails for the same reason.

**2. The corrected value could never deploy.** `ignoreCommand` skipped any redeploy where `VERCEL_GIT_PREVIOUS_SHA` equalled `HEAD` — exactly what an env-var edit triggers. The build reported *"Canceled by Ignored Build Step"*, which reads as success while the container kept running the old environment.

The open PR fixes `ignoreCommand`, strips trailing slashes in `main.go` as defense, and documents both traps in `CLAUDE.md`. It changes `api/`, so it produces a real diff and deploys normally — no deadlock.

**After it merges,** confirm the live value, then test login and course creation in production:

```bash
curl -sI https://surus-1.vercel.app/auth/google/callback | grep -i location
# want: https://surus-beta.vercel.app?error=auth_failed   (no slash before ?)
```

---

## Open work

| # | Task | Notes |
| - | ---- | ----- |
| 2.6 | Course elements — lesson tree, video/page/quiz subtypes | Never exercised. The 2026-08-13 pass covered sign-in, create, list, and search only. Do not infer from those |
| 2.8 | Refresh rotation | Implemented, never run. Needs a deliberate wait past the 900s access-token TTL, then confirm a new token is issued and a replayed refresh token is rejected |
| 2.10 | Magic-link email | Hash bug fixed, but `MagicLinkRequest` never sends an email — the Postmark call is a comment and `POSTMARK_*` is read by nothing. Until this is done, Google is the only working login |
| 2.7d | Logout doesn't clear the API-origin cookie | The API call is made from the Next server, so its deletion headers never reach the browser; the stateless JWT stays valid there for up to 900s. Refresh tokens and handoff codes *are* revoked, so the session cannot be extended. Option A erases this |
| 2.12 | Delete dead `POST /v1/auth/google/callback` | Mounted and live, accepts `state` without verifying it, no caller in `web/`. Its partner `GoogleStart` isn't routed at all. Delete rather than harden — needs your OK, since it removes a public endpoint |
| — | Linkrot job | `api/internal/jobs/linkrot.go` is an in-process 24h ticker, but the container scales to zero after 5 minutes idle, so it never fires. Make it an HTTP endpoint driven by Vercel Cron, or drop the feature |

---

## Open decisions

### 1 — Move the session to a custom domain (Option A)

**Current state: Option B shipped, deliberately, as a bridge.** The API sets `access_token` on its own host *and* mints a single-use 60s handoff code; the web app exchanges it server-side at `POST /v1/auth/handoff` and sets its own copy on the web origin.

The token lives in **two cookie jars on purpose.** Server components read the web origin's via `cookies()`; `apiFetch` runs in the browser and sends the API origin's. Fixing one alone breaks the other — do not "simplify" by removing either.

Cost of the bridge: two expiries to keep in sync, plus 2.7d. **Option A — `surus.com` for web, `api.surus.com` for the API, cookie scoped `Domain=.surus.com` — deletes the mechanism entirely**: one jar, one deletion, no handoff, and `SameSite=Lax` suffices. It also gives Google OAuth a permanent redirect URL.

Blocked only on buying a domain. A shared `Domain` is impossible today because both projects sit on `vercel.app`, which is on the Public Suffix List.

### 2 — Postmark account for magic-link email

Blocks 2.10. Without it, Google is the only way in.

---

## Environment gotchas

Each cost real time. All are invisible to git, since env files are ignored — which is exactly why "works on my machine" held for so long.

| Trap | Detail |
| ---- | ------ |
| `NEXT_PUBLIC_API_URL` needs the `/v1` suffix | Fetch layers concatenate `` `${API_URL}${path}` ``. Without it every request 404s in a way that looks like a login bug |
| `NEXT_PUBLIC_APP_URL` must have **no** trailing slash | Breaks CORS matching and doubles the slash in redirects. See `CLAUDE.md` |
| `DATABASE_URL` must be the Neon **pooled** URL | `-pooler` in the hostname. Cold starts churn connections; the direct URL exhausts limits |
| Google redirect URIs are per-environment | `http://localhost:8080/auth/google/callback` for local, plus the production one. Points at the **API**, carries no `/v1`, and plain `http` is valid for localhost. Missing it gives `redirect_uri_mismatch` |
| A stale shell keeps old `.env` values | A ping failure naming `user=surus database=surus_dev` after fixing `DATABASE_URL` means the shell predates the edit. Open a new one rather than re-debugging the URL |

---

## Deploy notes

Both Vercel projects track `main`. **Merging deploys straight to production — there is no staging gate.**

Preview deployments run on PRs but are **build verification only, not a test environment**: hostnames rotate per branch and Google OAuth only redirects to pre-registered URLs, so you cannot sign in on one. All functional testing is local, against a Neon branch.

**Schema is applied by hand — there is no migration runner.** Migrations are split deliberately: `001_initial_schema.up.sql` is safe to paste, `001_initial_schema.down.sql` drops every table. They shared one file until a `-- +goose Down` marker — meaningless to anything but goose — nearly dropped production on 2026-08-13. Adopting goose properly would remove the hazard; not yet done.

Stacked PRs do not survive squash or rebase merges: both rewrite SHAs and orphan the next branch's base. Use merge commits for stacks, or `git rebase --onto <old-base> <branch>` after each merge.

---

## Reference

- [Run any Dockerfile on Vercel](https://vercel.com/blog/dockerfile-on-vercel)
- [Deploying Git repositories with Vercel](https://vercel.com/docs/git)
- [Claude Code subagents — frontmatter reference](https://code.claude.com/docs/en/sub-agents)
