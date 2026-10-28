# T01 — Calculator domain: basic operations

- **Area:** backend
- **Depends on:** T00
- **Estimate:** 30 min
- **Status:** ✅ Done

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
2. **Division precision** — ✅ `DivRound(d2, resultScale)` with a domain constant `resultScale = 16`; the package-global `decimal.DivisionPrecision` is never mutated. Results that are not exact are rounded half away from zero to 16 decimal places (documented in the API docs).
3. **Input limits** — ✅ decimals are arbitrary precision, so a 1 000 000-digit operand would be valid but expensive. Reject operands with more than 50 significant digits or an exponent beyond ±1000 → `operand_too_large`. Limits are domain constants (`maxSignificantDigits = 50`, `maxExponent = 1000`).

## Acceptance criteria

- [x] `add ["2", "3"]` → `5`; `subtract ["5", "8"]` → `-3`; `multiply ["4", "2.5"]` → `10`; `divide ["10", "4"]` → `2.5`.
- [x] `add ["0.1", "0.2"]` → exactly `0.3`.
- [x] `divide ["1", "3"]` → `0.3333333333333333` (16 places).
- [x] `divide ["x", "0"]` → `division_by_zero` (including `["0", "0"]`).
- [x] `modulo` → `unsupported_operation`.
- [x] Binary op with 1 or 3 operands → `invalid_operand_count`.
- [x] `"abc"`, `""`, `"1,5"` → `invalid_operand`; `"1e3"` and `"-0.5"` accepted.
- [x] `divide ["2", "3"]` → `0.6666666666666667` (rounded half away from zero, not truncated).
- [x] 50-significant-digit operand accepted; 51-digit operand → `operand_too_large`.
- [x] `"1e1000"` and `"1e-1000"` accepted; `"1e1001"` and `"1e-1001"` → `operand_too_large`.
- [x] Results print without trailing zeros (`"10"`, not `"10.0"`).
- [x] Table-driven tests, one row per criterion; coverage ≥ 90% on `internal/domain/...`.

## Implementation notes

- **Exponent limit is the scientific exponent** (`d.ddd × 10^e`), so `"10e1000"` counts as `1e1001` → `operand_too_large`. Significant digits ignore leading and trailing zeros (`"000123000"` has 3).
- Zero is never too large (`"0e5000"` is just `0`).
- Exponents the library cannot represent at all (e.g. `"1e99999999999"`) fail parsing → `invalid_operand`.
- Errors carry request details in `Fields` (`operation`, `operand`, `expected`/`received`) via `WithFields`, which copies — sentinels are never mutated.

## Definition of done

`make check` green · ledger updated · commit `Completed: 'T01: Calculator domain'`.
