# T02 — Calculation use case and ports

- **Area:** backend
- **Depends on:** T01
- **Estimate:** 20 min
- **Status:** ✅ Done

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

1. ✅ **Supported-operations listing** (accepted: `Operations(ctx) []dto.OperationInfo{Name, Arity, Symbol}`, in display order) — should the port also expose `Operations(ctx) []OperationInfo` (name, arity, symbol) so the frontend can render from the backend? *Recommendation:* yes, small and makes T03's `GET /api/v1/operations` and T06's UI data-driven. Can be deferred if time is short.

## Acceptance criteria

- [x] Service scenarios mirror T01's (success + each error code passes through with `errors.Is`).
- [x] `Operands: ["2.50", "1"]` → result DTO operands `["2.5", "1"]`.
- [x] Cancelled context → returns `ctx.Err()` without calculating.
- [x] Mocks generated, not hand-written.
- [x] `docs/ARCHITECTURE.md` request-flow section updated.
- [x] Decision 1: `Operations(ctx)` lists add `+`, subtract `−`, multiply `×`, divide `÷` (arity 2) in display order.

## Implementation notes

- `calculator.Operations()` returns the registry in display order. `ParseOperation` looks names up in an index built from the same slice, so T04 adds each operation in one place.
- Operands are parsed before the arity check, so `["abc"]` for a binary operation reports `invalid_operand` rather than `invalid_operand_count`.
- `Operations(ctx)` returns no error: it reads a static registry and has nothing to cancel.

## Definition of done

`make check` green · ledger updated · commit `Completed: 'T02: Calculation use case'`.
