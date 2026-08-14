# Surus

A community-curated platform to create, fork, and share free courses. Surus lets anyone build structured courses with lessons and quizzes, fork existing courses to remix them, and discover content shared by others.

## Stack

| Layer    | Technology                          |
| -------- | ----------------------------------- |
| Frontend | Next.js 16, React 19, Tailwind v4, shadcn |
| Backend  | Go 1.26, chi router, sqlc + pgx     |
| Database | PostgreSQL on [Neon](https://neon.tech) |
| Auth     | Google OAuth, magic-link email      |
| Storage  | Cloudflare R2 (media uploads)       |

The API (`api/`) and the web app (`web/`) are **two independent services**. They share no build tooling and talk to each other only over HTTP.

## Quick Start

**Prerequisites:** Go 1.26+, Node.js 22+, pnpm 11, and access to the team's Neon project.

### 1. Install dependencies

```bash
pnpm install            # repo root — dev task runner
cd web && pnpm install  # web app
```

### 2. Point at a database

Surus runs on Neon. For local work, create a **branch** of the production database — a copy-on-write fork you can break freely — and use its **pooled** connection string. Pooled URLs have `-pooler` in the hostname; the direct URL will exhaust connection limits.

```bash
cp api/.env.example api/.env
cp web/.env.example web/.env.local
```

Then in `api/.env` set at minimum:

```bash
DATABASE_URL=postgresql://…-pooler.…neon.tech/neondb?sslmode=require
JWT_SECRET=            # any long random string for local use
GOOGLE_CLIENT_ID=      # from the team's Google Cloud project
GOOGLE_CLIENT_SECRET=
```

`web/.env.local` needs `NEXT_PUBLIC_API_URL=http://localhost:8080/v1`. **The `/v1` suffix is required** — every route except the OAuth redirects is mounted under it, and without it every request 404s in a way that looks like a login bug.

### 3. Register the OAuth redirect URI

In Google Cloud Console → APIs & Services → Credentials → the Surus OAuth client → **Authorized redirect URIs**, add exactly:

```
http://localhost:8080/auth/google/callback
```

Without this, signing in fails with `redirect_uri_mismatch`. Three things people get wrong:

- It points at the **API** on `:8080`, not the web app on `:3000`.
- No `/v1` — the OAuth redirect routes are mounted at the API root.
- Plain `http://` is correct. Google makes an explicit exception to its HTTPS rule for `localhost`.

Adding this does not affect the production URI. Both can be registered at once.

### 4. Apply the schema

There is **no migration runner** — schema changes are applied by hand.

A brand-new Neon branch needs the whole of `api/db/migrations/001_initial_schema.sql`. A branch forked from production already has the schema, so apply only statements added since the fork.

### 5. Start both services

```bash
pnpm dev
```

Runs the API on `:8080` and the web app on `:3000` in one terminal, with prefixed output. Ctrl-C stops both. `make up` is an alias for the same thing.

Open [http://localhost:3000](http://localhost:3000).

Individually, if you prefer separate terminals:

```bash
pnpm dev:api    # or: cd api && go run ./cmd/server
pnpm dev:web    # or: cd web && pnpm dev
```

The API must run from `api/` — `godotenv` reads `./.env` relative to the working directory.

## Architecture notes

**The two services are on different origins in every environment.** Locally that's `localhost:8080` and `localhost:3000`; deployed they're separate Vercel projects on separate domains. This single fact drives most of the auth complexity:

- Auth is a JWT in an `access_token` **cookie**, not a bearer header.
- Deployed, that cookie must be `SameSite=None; Secure` or browsers withhold it from cross-site requests. Locally it falls back to `Lax`, since browsers reject `None` without `Secure` over plain HTTP.
- CORS must echo a specific allowlisted origin, because credentials are enabled and `*` is invalid alongside them.
- Because the browser's cookie jars are per-origin, the deployed login flow hands the session to the web app through a short-lived single-use code rather than relying on a shared cookie. This is a temporary bridge; a shared parent domain would remove it.

When something works locally but fails deployed, start there.

**Testing:** there are no automated tests. Verify by building both sides (`go build ./... && go vet ./...`, `pnpm build && pnpm lint`) and exercising the actual page or endpoint.

Conventions for each side live next to the code: [`api/CLAUDE.md`](api/CLAUDE.md) and [`web/CLAUDE.md`](web/CLAUDE.md). Cross-cutting concerns and current work state are in [`CLAUDE.md`](CLAUDE.md) and [`PLAN.md`](PLAN.md).

## Deployment

Two Vercel projects backed by this one repo: **web** (root directory `web/`) and **api** (root directory `api/`, built as a container from `Dockerfile.vercel`).

Deployment is **production-only — there are no preview environments.** Preview hostnames rotate per branch and Google OAuth rejects unregistered redirect URLs, so a preview login could never complete. The consequence: **merging to the production branch deploys straight to production**, and GitHub Actions' build/lint gate is the only automated check in between. Verify locally first.

The API container is stateless and scales to zero after 5 minutes idle, so nothing may rely on in-process state or background timers.

## Optional: local Postgres

`docker-compose.yaml` still provides a local Postgres if you'd rather not use a Neon branch. Start it with `docker compose up -d` and point `DATABASE_URL` at it. You'll need to apply the schema yourself. Neon branches are preferred — they match production's schema and data exactly.
