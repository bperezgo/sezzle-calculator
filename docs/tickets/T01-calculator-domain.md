# T01 — Calculator domain: basic operations

- **Area:** backend
- **Depends on:** T00
- **Estimate:** 30 min
- **Status:** ⬜ Todo

## Goal

The pure business core: addition, subtraction, multiplication and division as value objects and rules, with exact decimal arithmetic and every edge case expressed as a domain error. No HTTP, no JSON.

## Scope

### In
- Dependency: [`github.com/shopspring/decimal`](https://github.com/shopspring/decimal) (v1.6.0). It is a pure library, so the domain may import it (depguard only bans platform/transport).
- `internal/domain/calculator`:
  - `Operation` value object (`add`, `subtract`, `multiply`, `divide`), parsed from a string, rejects unknown. Carries its **arity** (binary for these four) so unary ops in T04 fit without redesign.
  - `Operand` value object wrapping `decimal.Decimal`, built from the operand's **string** form (`decimal.NewFromString`) — never from `float64`, which would reintroduce binary rounding.
  - A `Calculation` built only through a factory (valid by construction) that evaluates to a `decimal.Decimal`.
- `internal/domain/error`: one `*Error{Kind, Code, Message, Fields}` type with sentinels and `errors.Is` by code (same pattern as the previous project). Codes: `unsupported_operation`, `invalid_operand`, `invalid_operand_count`, `division_by_zero`, `operand_too_large`.
- Kinds: `KindInvalidInput` (unusable request), `KindMath` (well-formed but mathematically undefined). No HTTP statuses in the domain.

### Out
- Service, ports, handlers (T02, T03). Power / sqrt / percentage (T04).

## Decisions

1. **Number type** — ✅ **`shopspring/decimal`** (developer's call): exact base-10 arithmetic, battle-tested, so `0.1 + 0.2 = 0.3` exactly. Consequences: operands and results travel as decimal **strings** across layers (see T02/T03); the frontend never does arithmetic on JS numbers.
2. **Division precision** — `decimal.Div` uses the package-global `DivisionPrecision` (16). *Recommendation:* call `DivRound(d2, resultScale)` with a domain constant `resultScale = 16` instead of mutating the global; document "results are rounded half away from zero to 16 decimal places when not exact".
3. **Input limits** — decimals are arbitrary precision, so a malicious operand with 1 000 000 digits is valid but expensive. *Recommendation:* reject operands with more than 50 significant digits or an exponent beyond ±1000 → `operand_too_large`.

## Acceptance criteria

- [ ] `add ["2", "3"]` → `5`; `subtract ["5", "8"]` → `-3`; `multiply ["4", "2.5"]` → `10`; `divide ["10", "4"]` → `2.5`.
- [ ] `add ["0.1", "0.2"]` → exactly `0.3`.
- [ ] `divide ["1", "3"]` → `0.3333333333333333` (16 places).
- [ ] `divide ["x", "0"]` → `division_by_zero` (including `["0", "0"]`).
- [ ] `modulo` → `unsupported_operation`.
- [ ] Binary op with 1 or 3 operands → `invalid_operand_count`.
- [ ] `"abc"`, `""`, `"1,5"` → `invalid_operand`; `"1e3"` and `"-0.5"` accepted.
- [ ] 51-digit operand → `operand_too_large`.
- [ ] Results print without trailing zeros (`"10"`, not `"10.0"`).
- [ ] Table-driven tests, one row per criterion; coverage ≥ 90% on `internal/domain/...`.

## Definition of done

`make check` green · ledger updated · commit `Completed: 'T01: Calculator domain'`.
