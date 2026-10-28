# Architecture

> Living document. T00 set up the skeleton; each ticket updates the sections it touches. T09 finalises it.

## Shape

```
 Browser ──► Next.js server (frontend/) ──HTTP/JSON──► Go API (backend/)
             Server Component page                     Gin router
             Server Action  ──► src/lib/api (server-only)   │
                                                            ▼
                                     platform/handler ──► domain/ports (driving port)
                                     (adapter)                 ▲
                                                               │ implemented by
                                                        domain/service ──► domain/calculator
                                                        (use case)         (value objects + pure rules)
```

- The **browser only talks to Next.js**. `BACKEND_URL` is a server-only variable, so there is no CORS, no backend URL in the bundle, and the form works before hydration (progressive enhancement).
- The **Go domain is pure**: no Gin, no `net/http`, no JSON. It is reached through a driving port (the use-case interface the HTTP handler depends on). There are no driven ports yet — no database, no external services, so no repositories.
- **Numbers are exact decimals** (`shopspring/decimal`) end to end: strings on the wire and in DTOs, `decimal.Decimal` inside the domain, never `float64` / JS `number`.
- **Composition roots:** `backend/cmd/main.go` (config → services → handlers → router → server) and the Next.js app router.

## Request flow

A calculation, from the HTTP adapter inward (T03 adds the handler; T02 provides everything below it):

1. `platform/handler` binds the JSON body into `dto.CalculationRequest{Operation, Operands []string}`. Operands stay strings, so no `float64` is ever involved.
2. It calls the driving port `ports.Calculator.Calculate(ctx, req)`. The handler only knows the interface; tests use the generated `mocks.MockCalculator`.
3. `service.CalculatorService` returns `ctx.Err()` at once if the request was cancelled. Otherwise it builds value objects: `calculator.ParseOperation` → `calculator.NewOperand` per operand → `calculator.NewCalculation` (arity check) → `Evaluate()`.
4. On success it returns `dto.CalculationResult{Operation, Operands, Result}`. Operands are echoed in canonical form (`"2.50"` → `"2.5"`) and the result is a canonical decimal string.
5. On failure the `*domainerror.Error` is returned **untouched** (`Kind`, `Code`, `Fields`). The handler-side mapper turns `Kind` into an HTTP status (T03).

`ports.Calculator.Operations(ctx)` lists the supported operations (`name`, `arity`, `symbol`) in display order, so the API (`GET /api/v1/operations`) and the UI are driven by the domain's registry.

**Persistence would add a driven port, not change this one.** Calculation history, for example, would be a `ports.CalculationRepository` (driven, implemented by an adapter in `platform/`) injected into `NewCalculatorService(repo)`. `ports.Calculator` and its callers would stay as they are.

## Backend packages

| Package | Holds | Imports allowed |
|---|---|---|
| `internal/domain/calculator` | Value objects (`Operation`, `Operand`, …) and pure arithmetic rules on `shopspring/decimal` | stdlib (no I/O), `shopspring/decimal`, `domain/error` |
| `internal/domain/dto` | Plain structs crossing layer boundaries; decimals travel as canonical strings | stdlib |
| `internal/domain/error` | Categorised domain errors (`Kind`, `Code`, `Message`) | stdlib |
| `internal/domain/ports` | Interfaces: the driving `Calculator` port (`Calculate`, `Operations`) (+ generated `mocks/`) | `domain/dto`, stdlib |
| `internal/domain/service` | Use cases implementing the ports | `domain/*` |
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
| `src/lib/` | `server-only` API client + shared API contract types, input validation, number formatting |

## Testing strategy

| Layer | Style | Doubles |
|---|---|---|
| `domain/calculator` | Table-driven unit tests over pure functions | none |
| `domain/service` | Scenario tests per spec requirement | none / mocks of driven ports if any appear |
| `platform/handler` | `httptest` + Gin test context, status + JSON body | mockery mock of the `Calculator` port |
| `platform/server` | End-to-end through the real router + real service | none |
| `frontend/src/lib` | Unit tests; `fetch` mocked with `vi.fn()` | mocked `fetch` |
| `frontend/src/components` | Testing Library through roles/labels, `user-event` | mocked Server Action |
| Stack | `docker compose up --wait` smoke test in CI | none |
