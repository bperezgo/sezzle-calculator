# T03 — REST API with Gin

- **Area:** backend
- **Depends on:** T02
- **Estimate:** 40 min
- **Status:** ⬜ Todo

## Goal

The HTTP adapter: endpoints for calculator operations that validate input, handle edge cases and always answer JSON.

## Scope

### In
- `platform/handler/calculation_handler.go`: bind JSON → validate shape → call `ports.Calculator` → encode.
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
4. **Number wire format** (follows from decimals in T01) — *Recommendation:* the **result is always a JSON string** (`"0.3"`), the convention for exact decimals in payment APIs, because JS clients would round a JSON number to float64. **Operands accept strings or JSON numbers** (`["0.1", 0.2]`) so curl examples stay friendly; numbers are read from their raw text (`json.Number`), never via `float64`.
5. **`GET /api/v1/operations`** — only if T02 decision 1 was accepted.

## Acceptance criteria

- [ ] `POST /api/v1/calculations {"operation":"add","operands":["2","3"]}` → 200 `"result":"5"`.
- [ ] `{"operation":"add","operands":[0.1, 0.2]}` (JSON numbers) → `"result":"0.3"` exactly.
- [ ] Divide by zero → 422 `division_by_zero`.
- [ ] Not JSON / empty body / `operands` not an array / an operand that is neither string nor number (`true`, `null`) → 400 `invalid_request`.
- [ ] `{"operation":"add","operands":["a","1"]}` → 400 `invalid_operand`.
- [ ] Unknown operation → 400 `unsupported_operation`; wrong arity → 400 `invalid_operand_count`.
- [ ] Every response (including 404/405/500) is `application/json`.
- [ ] Handler tests use the mockery mock; router tests go end-to-end with the real service.
- [ ] `X-Request-ID` present on every response.

## Definition of done

`make check` green · `make up` + curl examples from `docs/api/calculator.md` work · ledger updated · commit `Completed: 'T03: REST API'`.
