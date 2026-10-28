# T05 — Frontend API client and Server Action

- **Area:** frontend
- **Depends on:** T03
- **Estimate:** 30 min
- **Status:** ⬜ Todo

## Goal

The frontend's adapter to the Go API, running **only on the Next.js server**, plus the Server Action the UI will call.

## Scope

### In
- `src/lib/api/types.ts`: API contract types (`Operation`, `CalculationRequest`, `CalculationResponse`, `ApiError` with the error `code` union) — defined once. Operands and result are **strings** (exact decimals from T01/T03); the frontend never converts them to `number`.
- `src/lib/api/client.ts` (`import "server-only"`): `calculate(req)` → `fetch(`${BACKEND_URL}/api/v1/calculations`)` with a timeout (`AbortSignal.timeout`), `cache: "no-store"`, parses success and the error envelope, maps network/timeout/5xx to a typed `backend_unavailable` result. Returns a discriminated union (`{ ok: true, result } | { ok: false, error }`) instead of throwing.
- `src/lib/env.ts`: reads and validates `BACKEND_URL` once (fail fast with a clear message).
- `src/app/actions.ts` (`"use server"`): `calculateAction(prevState, formData)` for `useActionState` — parses/validates form fields, calls the client, returns serialisable state `{ status: "idle" | "success" | "error", ... }`.
- `frontend/.env.example` with `BACKEND_URL=http://localhost:8080`.

### Out
- UI (T06). Client-side `fetch` to the backend (never).

## Decisions

1. **Validation library** — *Recommendation:* no dependency; a small typed parser in `src/lib/validation.ts` is enough for two numbers and an operation enum. Alternative: `zod` (nicer errors, ~13 kB, but server-side only so bundle size doesn't matter much). Pick one.
2. **Read Next 16 docs first** — `node_modules/next/dist/docs/01-app/02-guides/server-actions.md` and `forms.md` (APIs may differ from training data; see `frontend/AGENTS.md`).

## Acceptance criteria

- [ ] Client unit tests with mocked `fetch`: success, each API error code, malformed JSON response, network error, timeout.
- [ ] Action tests: invalid form input never calls the client; valid input returns `success` state with the result.
- [ ] `BACKEND_URL` is not present in the client bundle (grep `.next/static` after `pnpm build`).

## Definition of done

`make check` green · ledger updated · commit `Completed: 'T05: Frontend API client'`.
