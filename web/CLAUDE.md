# web/ — Next.js frontend

Next.js 16 App Router, React 19, Tailwind v4, shadcn-style UI. Runs on `:3000`. Deploys as its own Vercel project with root directory `web/`.

**pnpm, never npm.**

```bash
pnpm dev
pnpm build
pnpm lint      # eslint flat config, eslint.config.mjs
```

No test runner exists. Verify by building and loading the page.

## Route structure

Route groups under `app/`:

- `(public)` — no auth required: landing, course view, lesson view, search, user profiles.
- `(auth)` — the group layout redirects unauthenticated users: dashboard, library, course create, course edit, settings.
- `admin/` — reports moderation.

## The two fetch layers

This is the single most common source of bugs here. Pick by component type:

**Server Components** → `lib/api/server/*.ts`, built on `serverFetch`.
- Forwards the `access_token` cookie manually via `cookies()` — you must pass `{ auth: true }` or the request goes out unauthenticated.
- Defaults to `cache: "no-store"`.
- Throws a plain `Error`.

**Client Components** → `lib/api/*.ts`, built on `apiFetch` (`lib/api/client.ts`).
- Sends `credentials: "include"`.
- Transparently retries once through `/auth/refresh` on a 401.
- Throws a typed `ApiError(code, message, status)`.
- Client data goes through TanStack Query. `useAuth` is the pattern to copy; `lib/providers.tsx` sets up the client.

A Client Component calling `serverFetch`, or a Server Component that never forwards the cookie, produces a silent 401 rather than an obvious crash. Check this first when data is missing for a logged-in user.

## Types

`lib/types.ts` mirrors the API JSON **by hand**. There is no codegen. Whenever a Go response struct changes, this file must change with it — otherwise a field silently reads as `undefined`. When a value is unexpectedly missing, compare this file against the Go `model.go` before assuming a fetch bug.

## UI conventions

- shadcn-style components in `components/ui/` (Radix primitives + `cn()` from `lib/utils.ts`).
- Tailwind v4 via `@tailwindcss/postcss`.
- Zustand (`lib/stores/uiStore.ts`) for **UI-only** state. Server state stays in TanStack Query — don't mirror fetched data into Zustand.
- Rich text is Tiptap. `components/editor/TiptapEditor.tsx` writes, `TiptapRenderer.tsx` reads. Content is Tiptap JSON, stored as-is by the API.

## Talking to the API

The API is a **separate Vercel project on a different origin** — never a same-origin `/api/*` route. `NEXT_PUBLIC_API_URL` must point at the right API deployment for the current environment; if it's unset or stale, every authenticated request fails in a way that looks like a login bug.

Because it's cross-origin, browser requests only carry the `access_token` cookie when the API sets `SameSite=None; Secure` and returns a CORS allowlist that includes this deployment's hostname. When debugging a failed request, check whether the cookie was actually sent before suspecting the component.
