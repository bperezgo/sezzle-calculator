# T02 — Calculation use case and ports

- **Area:** backend
- **Depends on:** T01
- **Estimate:** 20 min
- **Status:** ⬜ Todo

## Goal

Expose the domain through a **driving port** so the HTTP adapter depends on an interface, not on the domain internals. No repositories: there is no persistence.

## Scope

### In
- `internal/domain/dto`: `CalculationRequest{Operation string, Operands []string}`, `CalculationResult{Operation string, Operands []string, Result string}`. Decimal values cross layers as **canonical strings**, so DTOs stay stdlib-only and the domain owns parsing (and its `invalid_operand` error).
- `internal/domain/ports/calculator.go`: `type Calculator interface { Calculate(ctx context.Context, req dto.CalculationRequest) (dto.CalculationResult, error) }`.
- `internal/domain/service/calculator_service.go`: `NewCalculatorService() ports.Calculator` — maps DTO → value objects → result DTO; returns domain errors untouched. Result operands are echoed in canonical form (`"2.50"` → `"2.5"`).
- `make generate-mocks` → `ports/mocks/calculator.go` (used by handler tests in T03).
- Respect `ctx` cancellation (return early if `ctx.Err() != nil`).

### Out
- HTTP (T03). Driven ports — none needed now; note in `docs/ARCHITECTURE.md` that history/persistence would add a driven port, not change this one.

## Decisions

1. **Supported-operations listing** — should the port also expose `Operations(ctx) []OperationInfo` (name, arity, symbol) so the frontend can render from the backend? *Recommendation:* yes, small and makes T03's `GET /api/v1/operations` and T06's UI data-driven. Can be deferred if time is short.

## Acceptance criteria

- [ ] Service scenarios mirror T01's (success + each error code passes through with `errors.Is`).
- [ ] `Operands: ["2.50", "1"]` → result DTO operands `["2.5", "1"]`.
- [ ] Cancelled context → returns `ctx.Err()` without calculating.
- [ ] Mocks generated, not hand-written.
- [ ] `docs/ARCHITECTURE.md` request-flow section updated.

## Definition of done

`make check` green · ledger updated · commit `Completed: 'T02: Calculation use case'`.
