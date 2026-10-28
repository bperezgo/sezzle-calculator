# Sezzle Calculator — Agent Rules

Full-stack calculator (interview brief: `test.md`). Monorepo: **`backend/`** Go REST API (Gin, hexagonal + DDD value objects) and **`frontend/`** Next.js app (server-first, TypeScript, Tailwind). This file is the always-loaded rule set — keep it short. Read on demand:

- `docs/tickets/` — the plan. One ticket = one `Completed:` commit. Work them in order; the acceptance criteria are the spec.
- `docs/ARCHITECTURE.md` — layers, request flow, testing strategy.
- `docs/PROMPTS.md` — the AI prompt audit ledger (an interview deliverable).
- `frontend/AGENTS.md` — Next.js 16 rules: read `frontend/node_modules/next/dist/docs/` before writing Next code; APIs differ from training data.

## Commands

Root: `make check` (backend + frontend) · `make up` / `make down` (docker compose) · `make cover` · `make install-hooks`
Backend (`cd backend`): `make run` · `make test` · `make test-race` · `make cover` · `make lint` · `make fmt` · `make generate-mocks`
Frontend (`cd frontend`): `pnpm dev` · `pnpm test` · `pnpm test:coverage` · `pnpm lint` · `pnpm format` · `pnpm typecheck` · `pnpm build` · `pnpm check`

## Golden rules

- **Tickets are the spec.** Every acceptance criterion maps to at least one test, written before the code (name the test or table row after the criterion). If code and a criterion disagree, change the ticket deliberately — never drift silently. Tick criteria as they turn green.
- **Log prompts.** At the end of every task, append an entry to `docs/PROMPTS.md` with the developer's prompt **verbatim**, the ticket, and a short outcome. Never rewrite or delete earlier entries.
- **Domain never imports platform.** `backend/internal/domain/**` must not import `internal/platform`, Gin, `net/http` or `encoding/json` — enforced by depguard.
- **Browser never calls the backend.** Only server code in `frontend/` (Server Actions, Server Components, `server-only` modules) reads `BACKEND_URL` and calls the Go API.
- **No `any`** in Go domain signatures/ports/DTOs or in TypeScript (`interface{}`/`any` → concrete types, `unknown` + narrowing at boundaries).
- **Never hand-edit mocks** in `backend/internal/domain/ports/mocks/`; run `make generate-mocks`.
- **Ask before changing existing tests' expectations.** Bug fixes add a test that reproduces the bug.
- **Never read `.env`** files.
- **Done = `make check` green.** Run it before saying a task is complete.

## Backend conventions

- `context.Context` first param on every service, port and adapter method.
- Constructors `NewXxx(deps...)`; adapters return the **port interface**. Naming `XxxService`, `XxxHandler`; no `Interface` suffix.
- Errors: categorised domain errors in `domain/error` with a `Kind` and a stable `Code`; the domain never knows HTTP statuses — one handler-side mapper does. The domain never logs.
- API error codes/messages are **contract**: constants, asserted verbatim in tests.
- Comments: default to none; only a non-obvious *why* in 1–2 lines.

## Frontend conventions

- Server Components by default; `"use client"` only for the interactive leaf (the calculator form/keypad).
- Calls to the backend live in `src/lib/` behind `import "server-only"`; types for the API contract are defined once there.
- Validate input on the client for UX and again on the server (Server Action) before calling the API; the backend is the final authority.
- Accessible by default: labelled inputs, `aria-live` for results/errors, keyboard usable.

## Testing

- Go: testify (`assert`, `require`, `mock`), table-driven; name `Test{Type}_{Method}_{Scenario}`; one test/row per spec scenario. Target ≥90% on `domain/`.
- Frontend: Vitest + Testing Library; test behaviour through roles/labels, not implementation details. Mock `fetch` at the `src/lib` boundary.

## Commits

- One commit series per ticket; the last is titled `Completed: 'Tnn: <Ticket title>'`. Small focused commits in between are fine. Never commit with `make check` red.
