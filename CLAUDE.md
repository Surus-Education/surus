# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Surus: community platform to create, fork, and share free courses. Go/chi API (`api/`) + Next.js App Router frontend (`web/`) + PostgreSQL on Neon. Two separate services — no monorepo tooling links them; the frontend talks to the API over HTTP only.

Detailed conventions live next to the code:

- **`api/CLAUDE.md`** — package layout, sqlc workflow, error contract, auth, domain model.
- **`web/CLAUDE.md`** — route groups, the two fetch layers, types mirroring, UI conventions.

Read the one for the directory you're working in. This file covers what spans both.

## Agent handoff protocol

`PLAN.md` is the shared context handoff between agents. It is the task board; the `CLAUDE.md` files are the standing knowledge. Keep the split — durable facts about how the code works belong in a `CLAUDE.md`, current work state belongs in `PLAN.md`.

**On start**, before touching anything:

1. Read `PLAN.md` in full, plus the scoped `CLAUDE.md` for the directory you own.
2. Find your task. Set its status to `in progress` and put your agent name in the Owner column.
3. If the task's premise looks wrong, say so in your report rather than silently redefining scope.

**While working:**

- Anything you discover that a later agent would otherwise have to rediscover goes in `PLAN.md` — a wrong assumption, a real root cause, a dead end that cost time. Add rows to **Known broken** as you confirm symptoms.
- A question you cannot resolve yourself becomes a row in **Open decisions**, with what it blocks. Do not guess and move on.
- Stay inside your directory scope. Cross-boundary work gets a new subtask for the agent that owns that side.

**On finish:**

1. Set the task status to `done`, or `blocked` with the reason inline.
2. Append to **Completed**: one paragraph — what changed, which files, and why that approach. Enough that a cold agent understands the shape without reading the diff.
3. Clear any Open decisions you resolved, recording the answer.
4. Update `Last updated`.

Report failures plainly. A task that is half done is `in progress` with a note, never `done`.

## Response style — caveman mode

Chat responses and agent report summaries use **caveman mode**: drop articles, filler
(just/really/basically/actually/simply), pleasantries, and hedging. Fragments are fine. Prefer short
synonyms. Technical terms, symbol names, and quoted errors stay exact. Pattern:
`[thing] [action] [reason]. [next step].`

This applies always, in every session and every agent, without being asked. It is off only when
someone says "stop caveman" or "normal mode".

**Written artifacts stay normal prose.** Caveman is a conversation register, not a documentation
style. Write in full sentences in:

- code and code comments
- commit messages and PR descriptions
- `PLAN.md` and the `CLAUDE.md` files

`PLAN.md` exists so a cold agent can pick up work without re-deriving context. Compressing it
defeats its only purpose. Same for code comments — they outlive the conversation.

Also drop caveman for security warnings, confirmations of irreversible actions, and any multi-step
sequence where dropped conjunctions would make the order ambiguous. Resume after.

## Agents

The agent definitions themselves (`.claude/agents/*.md`) and the MCP wiring (`.mcp.json`) are
**gitignored** — they encode one person's local tool installation, not a team convention. The roster
below is documentation of how work is divided, not a pointer to files you'll find in a fresh clone.
Nothing in this repo depends on these agents existing; the directory-ownership rules apply to whoever
is editing, human or otherwise.

| Agent      | Model      | Owns              | Use for |
| ---------- | ---------- | ----------------- | ------- |
| `backend`  | Sonnet 5   | `api/`            | handlers, services, sqlc, migrations |
| `frontend` | Sonnet 5   | `web/`            | pages, components, fetch layers, UI |
| `debug`    | Sonnet 5   | nothing (reports) | reproducing failures, root-causing |
| `planner`  | Opus 4.8   | `PLAN.md`         | decomposing work, keeping the board honest |
| `lookup`   | Haiku 4.5  | nothing           | finding code, reading library docs |

`backend` and `frontend` may not edit each other's directory. Work spanning both becomes two subtasks with a stated dependency.

## Commands

```bash
pnpm dev                        # both services, one terminal (root package.json)
pnpm dev:api                    # API alone, :8080
pnpm dev:web                    # web alone, :3000

cd api && go build ./... && go vet ./...
cd web && pnpm build
cd web && pnpm lint
```

`pnpm dev` runs both through `concurrently`, which spawns the processes directly. The old `make up` shelled out via `nohup sh -c`, which is broken on Windows — make hands the command to a Unix shell whose PATH has no Windows `go` or `node`. `make up` is now a thin alias for `pnpm dev`.

The API must run from `api/` — `godotenv` reads `./.env` relative to the working directory. The `dev:api` script handles this.

There are **no tests** in this repo — no `_test.go` files, no test runner in `web/package.json`. Verify changes by building both sides and exercising the endpoint or page. Say explicitly when you have not exercised it.

Env: copy `api/.env.example` → `api/.env` and `web/.env.example` → `web/.env.local`.

Database is **Neon**. For local work, point `DATABASE_URL` at a Neon branch — a disposable copy-on-write fork of production — rather than the Docker Postgres in `docker-compose.yaml`, which remains available as a fallback.

## Deploy

Two Vercel projects backed by this one repo:

- **web** — root directory `web/`, standard Next.js build.
- **api** — root directory `api/`, deployed as a container. Vercel detects `Dockerfile.vercel` at the project root, builds the image, pushes it to the Vercel Container Registry, and serves it from a Fluid-compute Function.

Deployment is driven by Vercel's native Git integration. Merging to `main` deploys straight to production — there is no staging gate.

**Vercel preview deployments do run on pull requests, but they are build verification only, not a test environment.** A preview proves the project compiles and deploys; it cannot be used to exercise the app. Preview hostnames rotate per branch, and Google OAuth only redirects to pre-registered URLs, so login can never complete on one. Everything downstream of login — which is nearly the whole app — is unreachable there.

The practical rule: **all functional testing happens locally**, against a Neon branch. A green preview means "it builds," nothing more. Production is the first environment where the deployed cross-origin behavior is genuinely exercised, which is why that path deserves a real check immediately after a merge.

GitHub Actions runs build and lint gates only; it never deploys.

### `vercel.json` — the ignoreCommand, and why it looks like that

Both projects share one repo, so each skips builds for commits that changed nothing in its own root directory. That is the `ignoreCommand` in `web/vercel.json` and `api/vercel.json`:

```sh
base=$(git rev-parse -q --verify "${VERCEL_GIT_PREVIOUS_SHA:-HEAD~1}^{commit}") && git diff --quiet "$base" HEAD -- . || exit 1
```

Two rules govern it, and both have already broken a deployment:

**Only exit codes 0 and 1 mean anything.** 0 skips the build, 1 builds. Anything else fails the deployment outright rather than falling back. A plain `git diff` against `VERCEL_GIT_PREVIOUS_SHA` exits 128 when that object is missing from the build clone — which happens after any force-push, and can happen in a shallow clone. Hence verifying the base first and falling through to `exit 1`: **uncertainty means build.** A redundant build is cheap; a failed check blocks the PR.

**`vercel.json` rejects unknown keys.** The schema is strict (`should NOT have additional property`) and JSON has no comment syntax, so a `_comment` field fails validation. That is why this explanation lives here instead of next to the command.

**The two services are on different origins in every deployed environment.** That single fact drives most of the cross-cutting complexity: the `access_token` cookie is cross-site (`SameSite=None; Secure`), CORS must echo a specific allowlisted origin because credentials are enabled, and `NEXT_PUBLIC_API_URL` on the web side must point at the matching API deployment. When something works locally and fails deployed, start there.

The API container is **stateless and scales to zero after 5 minutes idle**. Nothing may rely on in-process state or long-lived background timers.

## Branches

All feature code is on `main`. `vercel` is `main` plus deploy experiments. `list` is identical to `main`; `origin/eric/course-create` duplicates a commit already on `main`. **Neither holds unmerged features** — if something is missing, it is a bug, not a merge.

Current work state, open decisions, and known breakage: see `PLAN.md`.
