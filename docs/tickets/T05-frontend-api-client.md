# T05 — Frontend API client and Server Action

- **Area:** frontend
- **Depends on:** T03
- **Estimate:** 30 min
- **Status:** ✅ Done

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

1. **Validation library** — ✅ **no dependency**; a small typed parser in `src/lib/validation.ts` (`parseCalculationForm`, `isDecimal`) is enough for two numbers and an operation enum. It returns field error *codes* (`required`, `invalid_number`, `unsupported_operation`); T07 maps them to messages. Alternative considered: `zod` (nicer errors, ~13 kB, server-side only).
2. **Read Next 16 docs first** — ✅ read `server-actions.md`, `forms.md`, `instrumentation.md` and the `server-only` section of Server and Client Components.
3. **Form fields and state** — ✅ fields `operation`, `a`, `b` (the API's `[a, b]`). The action state (`src/lib/calculator-state.ts`, outside `actions.ts` because a `"use server"` module may only export async functions) always echoes `values` so inputs survive errors; `error` state carries `fieldErrors` (validation) and/or `error` (API).
4. **Contract violations** — ✅ a 2xx/4xx body that doesn't match the contract (bad JSON, missing/non-string `result`, unknown operation or error code) is mapped to `backend_unavailable`, like network/timeout/5xx: the UI can't trust it, and the user's action is the same (try again).
5. **Fail fast** — ✅ `src/instrumentation.ts` calls `backendUrl()` at server start, so a missing/invalid `BACKEND_URL` is reported before the first request. `next build` doesn't need it (read at runtime, never inlined). Tests alias `server-only` to its empty module; Next enforces the boundary in the build.

## Acceptance criteria

- [x] Client unit tests with mocked `fetch`: success, each API error code, malformed JSON response, network error, timeout. (`src/lib/api/client.test.ts`)
- [x] Action tests: invalid form input never calls the client; valid input returns `success` state with the result. (`src/app/actions.test.ts`)
- [x] `BACKEND_URL` is not present in the client bundle (grep `.next/static` after `pnpm build`). Built with `BACKEND_URL=http://backend-sentinel.invalid:8080`: neither the name nor the value appears in `.next/static` (nor in `.next/server`, since it's read at runtime). Re-check in T06 once the page imports the action.

## Definition of done

`make check` green · ledger updated · commit `Completed: 'T05: Frontend API client'`.
