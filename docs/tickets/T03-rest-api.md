# T03 — REST API with Gin

- **Area:** backend
- **Depends on:** T02
- **Estimate:** 40 min
- **Status:** ✅ Done

## Goal

The HTTP adapter: endpoints for calculator operations that validate input, handle edge cases and always answer JSON.

## Scope

### In
- `platform/handler/calculation_handler.go`: bind JSON → validate shape → call `service.CalculatorService` → encode.
- `platform/handler/error_response.go`: the **single** domain-error → HTTP mapper (`KindInvalidInput` → 400, `KindMath` → 422, unknown → 500 with a generic message, cause logged not shown).
- Malformed JSON / wrong types / missing fields → 400 `invalid_request`; body over 1 MiB → 413; unknown route → 404 JSON; wrong method → 405 JSON.
- `platform/middleware`: request ID (`X-Request-ID` echo/generate), `slog` access log, JSON panic recovery.
- Routes in `platform/server/router.go`; wiring in `cmd/main.go`.
- `docs/api/calculator.md` with request/response examples (curl).

### Out
- CORS — not needed: only the Next.js server calls the API (see `docs/ARCHITECTURE.md`). Auth, rate limiting.

## Decisions

1. **Endpoint shape** — ✅ one resource, `POST /api/v1/calculations` with `{"operation":"divide","operands":["10","2"]}` → `200 {"operation":"divide","operands":["10","2"],"result":"5"}`. Unary ops (sqrt) fit through operand arity; one handler/test suite covers every op.
2. **Error envelope** — ✅ `{"error":{"code":"division_by_zero","message":"division by zero is undefined"}}`; `code` is the stable contract the frontend switches on, `message` is human-readable.
3. **Status for math errors** — ✅ 422 (well-formed request, undefined math); 400 for unusable requests.
4. **Number wire format** (follows from decimals in T01) — ✅ accepted as recommended: the **result is always a JSON string** (`"0.3"`), the convention for exact decimals in payment APIs, because JS clients would round a JSON number to float64. **Operands accept strings or JSON numbers** (`["0.1", 0.2]`) so curl examples stay friendly; numbers are read from their raw text (`json.Number`), never via `float64`.
5. **`GET /api/v1/operations`** — ✅ T02 decision 1 was accepted: `200 {"operations":[{"name":"add","arity":2,"symbol":"+"},…]}` in display order.

## Acceptance criteria

- [x] `POST /api/v1/calculations {"operation":"add","operands":["2","3"]}` → 200 `"result":"5"`.
- [x] `{"operation":"add","operands":[0.1, 0.2]}` (JSON numbers) → `"result":"0.3"` exactly.
- [x] Divide by zero → 422 `division_by_zero`.
- [x] Not JSON / empty body / `operands` not an array / an operand that is neither string nor number (`true`, `null`) → 400 `invalid_request`.
- [x] `{"operation":"add","operands":["a","1"]}` → 400 `invalid_operand`.
- [x] Unknown operation → 400 `unsupported_operation`; wrong arity → 400 `invalid_operand_count`.
- [x] Every response (including 404/405/500) is `application/json`.
- [x] Acceptance tests in `backend/tests/` drive the fully wired router with the real service (no mocks). *Changed after review: the `ports.Calculator` interface and its mockery mock were removed as unneeded indirection, and the handler/router end-to-end tests moved to `backend/tests/`.*
- [x] `X-Request-ID` present on every response.

- [x] Decision 5: `GET /api/v1/operations` lists the domain registry in display order.

## Implementation notes

- Transport codes are contract constants in `handler/error_response.go`: `invalid_request`, `payload_too_large` (413), `not_found`, `method_not_allowed` (405, with `Allow`), `internal_error`.
- `operation`/`operands` missing or `null` → `invalid_request`; `"operation":""` or `"operands":[]` reach the domain (`unsupported_operation` / `invalid_operand_count`). Trailing data after the JSON object → `invalid_request`. The `Content-Type` header is not required, so plain `curl -d` works.
- Unknown errors are recorded with `c.Error` and logged by the access log at ERROR with the request ID; panics are logged with their stack by `middleware.Recovery`. Neither reaches the client.
- Middleware order: request ID → access log → recovery → 1 MiB body limit, so even panics carry `X-Request-ID` and are logged as 500s. Incoming request IDs outside `[A-Za-z0-9._-]{1,128}` are replaced (log-injection guard).
- The error envelope has no `details` yet; domain `Fields` (e.g. expected/received arity) stay server-side.

## Definition of done

`make check` green · `make up` + curl examples from `docs/api/calculator.md` work · ledger updated · commit `Completed: 'T03: REST API'`.
