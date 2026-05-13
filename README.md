# Surus

A community-curated platform to create, fork, and share free courses. Surus lets anyone build structured courses with lessons and quizzes, fork existing courses to remix them, and discover content shared by others.

## Stack

| Layer    | Technology                          |
| -------- | ----------------------------------- |
| Frontend | Next.js 15, React, Tailwind, shadcn |
| Backend  | Go, chi router                      |
| Database | PostgreSQL 15                       |
| Auth     | Google OAuth, magic-link email      |
| Storage  | Cloudflare R2 (media uploads)       |

## Quick Start

**Prerequisites:** Go, Node.js, pnpm, Docker

### 1. Start the database

```bash
docker compose up -d
```

### 2. Configure environment

Copy the example env files and fill in any missing values:

```bash
cp api/.env.example api/.env      # set JWT_SECRET, Google OAuth credentials, etc.
cp web/.env.example web/.env.local
```

The defaults in `api/.env` connect to the local Docker postgres instance on port `5432`. The API runs on `:8080` and the frontend on `:3000`.

### 3. Start everything

```bash
make up
```

This starts the API and frontend in the background. Logs are written to `.api.log` and `.web.log`.

Open [http://localhost:3000](http://localhost:3000).

### Stop everything

```bash
make down
```

## Development

Run services individually if you prefer:

```bash
# API
cd api && go run ./cmd/server

# Frontend
cd web && pnpm dev
```
