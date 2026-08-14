# api/ — Go API

Go 1.26 + chi + pgx/sqlc. Runs on `:8080` locally, deploys as a Vercel container (`Dockerfile.vercel`).

Run it from `api/` — `godotenv` loads `.env` from the working directory.

```bash
go run ./cmd/server
go build ./...
go vet ./...
```

No tests exist. Verify by building and exercising the endpoint.

## Package layout

Feature packages live in `internal/<feature>/` and follow handler / service / model / routes:

- **routes.go** — returns a `chi.Router` mounted by `cmd/server/main.go`. Convention: `r.Use(authCfg.OptionalAuth)` for reads, then an `r.Group` with `authCfg.RequireAuth` for writes.
- **handler.go** — parse and validate (`go-playground/validator` on struct tags in model.go), call the service, respond. Handlers don't touch the DB, except in the thin packages (`quiz`, `fork`, `report`) which hold `*pgxpool.Pool` + `*db.Queries` directly.
- **service.go** — holds `pool` + `db.New(pool)`; returns `*middleware.ServiceError` for anything the client should see.
- **model.go** — request/response structs with `json` + `validate` tags.

## sqlc

`db/*.sql.go` is **generated — never edit by hand.**

1. Edit `db/queries/*.sql` (queries) or `db/migrations/001_initial_schema.sql` (schema).
2. Run `sqlc generate` **from the repo root** — `sqlc.yaml` lives there.
3. Apply the schema change yourself. There is **no migration runner**; the server only connects and pings.

Apply schema changes on a Neon branch first, verify, then promote. Never apply an unverified change to the production branch.

Type overrides in `sqlc.yaml`: `uuid` → `google/uuid.UUID`, `timestamptz` → `time.Time`, `jsonb` → `json.RawMessage` (nullable), `text[]` → `[]string`. Nullable columns without an override surface as `pgtype.*` — construct them as `pgtype.UUID{Bytes: id, Valid: true}`.

## Errors and responses

Everything client-visible goes through `middleware.ServiceError` (`internal/middleware/errors.go`). Services return `middleware.NewServiceError(code, msg)` or `WrapServiceError(code, msg, err)`; handlers call `middleware.HandleServiceError(w, r, err)`, which maps the code to an HTTP status via `statusMap`.

Valid codes: `validation_error`, `unauthenticated`, `forbidden`, `not_found`, `conflict`, `unprocessable`, `rate_limited`, `internal_error`.

Anything that isn't a `ServiceError` is logged and returned as a generic 500. **Never let a raw DB error reach the response.**

Success responses use a named envelope: `{"course": ...}`, `{"data": [...], "next_cursor": null}`. Errors are `{"error": {"code", "message"}}`.

## Auth

JWT in an **`access_token` cookie**, not a bearer header — `AuthConfig.extractUser` reads the cookie only. Claims: `sub` (uuid), `email`, `is_admin`.

Middlewares: `RequireAuth`, `RequireAdmin`, `OptionalAuth`. `OptionalAuth` attaches the user if present and never rejects, so `middleware.UserFromContext(ctx)` returns `nil` under it — always nil-check.

Cookies are set in `auth.Service.SetTokenCookies`. In every deployed environment the web app is on a **different origin** than the API, so the cookie is cross-site: it needs `SameSite=None` with `Secure`. Local development over plain HTTP can't use `SameSite=None` (browsers reject it without `Secure`), so the mode is chosen from `APP_ENV`. Don't hardcode one mode.

CORS lives in `middleware.CORS`, driven by `NEXT_PUBLIC_APP_URL` plus the comma-separated `CORS_ALLOWED_ORIGINS`. Because credentials are allowed, the response must echo a **specific** origin — `*` is invalid alongside `Access-Control-Allow-Credentials: true`, and the browser would drop the cookie.

Entries support a single `*` wildcard, but deployment is production-only so exact origins are expected. Don't introduce a wildcard without a concrete reason: a broad pattern like `https://*.vercel.app` would let any site on that domain make credentialed requests against this API.

Google OAuth browser redirects are mounted at **root** (`/auth/google/start`, `/auth/google/callback`), not under `/v1`, because the registered Google redirect URL has no prefix. Everything else is under `/v1`.

Rate limits: 300 req global, 10 req on `/v1/auth/*`.

## Domain model

- **Lessons are a tree.** `lessons.parent_id` self-references; ordering is `position` within `(course_id, parent_id)`. `CreateLesson` computes the next position via `GetMaxPosition` when none is given.
- **Lesson subtypes** are 1:1 tables keyed by `lesson_id`: `video_lessons`, `page_lessons`, `quiz_lessons`. Create/update writes the base row *then* switches on `type` for the subtype row; `ListLessons`/`GetLesson` stitch them into one `LessonResponse`. Adding a type touches the enum, a new table, queries, the switch in `lesson/service.go`, `web/lib/types.ts`, and a renderer in `web/components/lesson/`.
- **Rich text is Tiptap JSON** in `jsonb` (`page_lessons.content`, `video_lessons.curator_notes`, `quiz_lessons.questions`), passed through as `json.RawMessage` and never parsed server-side.
- **Forking** copies a course and its whole lesson tree in one transaction (`internal/fork/handler.go`), setting `forked_from_id` / `forked_at`.
- **Search** uses a `tsvector` column on `courses` maintained by trigger (title A / description B / tags C), queried with `websearch_to_tsquery`.
- **Visibility** is `public | unlisted | private`. `GetCourseByID` takes a `ViewerID` so private courses resolve for their owner and 404 for everyone else — pass the viewer through from `OptionalAuth` rather than filtering afterward.

## Deployment constraints

The container runs as a Vercel Function on Fluid compute and **scales to zero after 5 minutes idle**. It is stateless with no durable local storage.

- Don't add in-process background timers or rely on process memory surviving between requests. `internal/jobs/linkrot.go` currently violates this — its 24h ticker never fires in production. Tracked in `PLAN.md`.
- Use the Neon **pooled** connection string. Cold starts churn connections and the direct URL will exhaust limits.
