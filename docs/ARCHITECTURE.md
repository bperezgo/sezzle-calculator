# Architecture

> Living document. T00 set up the skeleton; each ticket updates the sections it touches. T09 finalises it.

## Shape

```
 Browser ──► Next.js server (frontend/) ──HTTP/JSON──► Go API (backend/)
             Server Component page                     Gin router
             Server Action  ──► src/lib/api (server-only)   │
                                                            ▼
                                     platform/handler ──► domain/service ──► domain/calculator
                                     (adapter)            (use case)         (value objects + pure rules)
```

- The **browser only talks to Next.js**. `BACKEND_URL` is a server-only variable, so there is no CORS, no backend URL in the bundle, and the form works before hydration (progressive enhancement).
- The **Go domain is pure**: no Gin, no `net/http`, no JSON. The HTTP handler calls the use case (`service.CalculatorService`) directly: with one implementation and no I/O, an interface would only add indirection, and the real service is cheap to use in handler tests. No database, no external services, so no repositories.
- **Numbers are exact decimals** (`shopspring/decimal`) end to end: strings on the wire and in DTOs, `decimal.Decimal` inside the domain, never `float64` / JS `number`.
- **Composition roots:** `backend/cmd/main.go` (config → services → handlers → router → server) and the Next.js app router.

## Request flow

A calculation, from the HTTP adapter inward:

0. `platform/server` runs every request through `middleware`: request ID (`X-Request-ID` echoed or generated) → `slog` access log → JSON panic recovery → 1 MiB body limit. Unknown routes/methods answer the JSON error envelope (404/405).

1. `platform/handler` decodes the JSON body into `dto.CalculationRequest{Operation, Operands []string}`. Operands may be JSON strings or numbers; numbers keep their raw text, so no `float64` is ever involved. Unusable bodies stop here with 400 `invalid_request` (413 over the limit).
2. It calls `service.CalculatorService.Calculate(ctx, req)` directly; handler tests use the real service.
3. `service.CalculatorService` returns `ctx.Err()` at once if the request was cancelled. Otherwise it builds value objects: `calculator.ParseOperation` → `calculator.NewOperand` per operand → `calculator.NewCalculation` (arity check) → `Evaluate()`.
4. On success it returns `dto.CalculationResult{Operation, Operands, Result}`. Operands are echoed in canonical form (`"2.50"` → `"2.5"`) and the result is a canonical decimal string.
5. On failure the `*domainerror.Error` is returned **untouched** (`Kind`, `Code`, `Fields`). `handler/error_response.go` is the single mapper: `KindInvalidInput` → 400, `KindMath` → 422, anything else → 500 `internal_error` with the cause logged, never returned. See `docs/api/calculator.md` for the contract.

`CalculatorService.Operations(ctx)` lists the supported operations (`name`, `arity`, `symbol`) in display order, so the API (`GET /api/v1/operations`) and the UI are driven by the domain's registry.

**Interfaces arrive with real dependencies.** Calculation history, for example, would add a `CalculationRepository` interface defined next to the service (implemented by an adapter in `platform/`, faked in service tests) and injected into `NewCalculatorService(repo)`. The handler would not change.

## Backend packages

| Package | Holds | Imports allowed |
|---|---|---|
| `internal/domain/calculator` | Value objects (`Operation`, `Operand`, …) and pure arithmetic rules on `shopspring/decimal` | stdlib (no I/O), `shopspring/decimal`, `domain/error` |
| `internal/domain/dto` | Plain structs crossing layer boundaries; decimals travel as canonical strings | stdlib |
| `internal/domain/error` | Categorised domain errors (`Kind`, `Code`, `Message`) | stdlib |
| `internal/domain/service` | Use cases (`CalculatorService`: `Calculate`, `Operations`) | `domain/*` |
| `internal/platform/handler` | Gin handlers: bind → validate → call port → encode; the single error → HTTP mapper | `domain/*`, Gin |
| `internal/platform/server` | Router (route table + middleware) | `platform/*`, Gin |
| `internal/platform/middleware` | Request ID, access log (`slog`), recovery, body limit | Gin, stdlib |
| `internal/platform/config` | Env configuration | stdlib |

The dependency rule is enforced by golangci-lint **depguard**: anything under `internal/domain` importing `internal/platform`, Gin, `net/http` or `encoding/json` fails `make lint`.

## Frontend layout

| Path | Holds |
|---|---|
| `src/app/` | Routes. `page.tsx` is a Server Component; `actions.ts` holds Server Actions |
| `src/components/` | UI. Only the interactive calculator is a Client Component |
| `src/lib/` | `server-only` API client (`api/client.ts`) + shared API contract types (`api/types.ts`), `env.ts` (`BACKEND_URL`), input validation, action state, number formatting |
| `src/instrumentation.ts` | Validates `BACKEND_URL` once at server start (fail fast) |

## Testing strategy

| Layer | Style | Doubles |
|---|---|---|
| `domain/calculator` | Table-driven unit tests over pure functions | none |
| `domain/service` | Scenario tests per spec requirement | none |
| `platform/handler`, `platform/middleware` | Unit tests for what HTTP cannot reach or isolates better (mapper fallback, request-ID rules, log shape) | none |
| `backend/tests/` | **Acceptance**: black-box HTTP through the fully wired router + middleware + real service; one test/row per ticket criterion | none |
| `frontend/src/lib` | Unit tests; `fetch` mocked with `vi.fn()` | mocked `fetch` |
| `frontend/src/components` | Testing Library through roles/labels, `user-event` | mocked Server Action |
| Stack | `scripts/smoke.sh` (`make smoke`, CI `docker` job) against `docker compose up --wait`: real API calculations and the no-JS form post through the backend | none |
