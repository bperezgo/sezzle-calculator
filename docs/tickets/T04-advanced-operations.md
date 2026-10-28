# T04 — Advanced operations: power, square root, percentage

- **Area:** backend
- **Depends on:** T03
- **Estimate:** 25 min
- **Status:** ✅ Done

## Goal

The brief's optional operations, added purely in the domain — no handler or router change should be needed (proves the design).

## Scope

### In
- `power` (binary) via `decimal.PowWithPrecision(exp, resultScale)` — it returns errors for the undefined cases instead of a silent `0`, which plain `Pow` does.
- `sqrt` (unary) via `decimal.SqrtRound(resultScale)`.
- `percentage` (binary): `a% of b = a × b / 100`.
- New error codes: `negative_square_root`, `undefined_power` (`0 ^ 0`, `0 ^ -1`, negative base with fractional exponent), `result_too_large`.
- Guard against resource exhaustion: exponent magnitude ≤ 1000 and integer exponent results capped (e.g. ≤ 1000 digits) → `result_too_large`. Arbitrary precision means `10 ^ 1e9` would otherwise try to allocate it.
- Update `docs/api/calculator.md`.

### Out
- Trigonometry, logarithms, expression parsing.

## Decisions

1. **Percentage semantics** — ✅ `percentage(a, b) = a% of b = a × b / 100` (`["20", "50"]` → `10`).
2. **Square root of negative** — ✅ error (`negative_square_root`), no complex numbers.
3. **`0 ^ 0`** — mathematically debated; `decimal` treats it as undefined. ✅ follow the library → `undefined_power` (simpler and documented), rather than the `1` convention.
4. **Size guard** — ✅ `|exponent|` ≤ 1000, and a power result must stay within the operand bounds' exponent (±1000) and have ≤ 1000 significant digits; otherwise `result_too_large`. Out-of-range cases are rejected from an estimate before any big-number work, then checked exactly. All three new codes are `KindMath` → 422.

## Acceptance criteria

- [x] `power ["2", "10"]` → `1024`; `power ["2", "-1"]` → `0.5`; `power ["4", "0.5"]` → `2`.
- [x] `power ["0", "-1"]`, `power ["0", "0"]`, `power ["-8", "0.5"]` → `undefined_power`.
- [x] `power ["10", "100000"]` → `result_too_large` (fast, no huge allocation).
- [x] `sqrt ["16"]` → `4`; `sqrt ["2"]` → `1.414213562373095` (rounded to 16 places, trailing zero trimmed); `sqrt ["-1"]` → `negative_square_root`; `sqrt ["1", "2"]` → `invalid_operand_count`.
- [x] `percentage ["20", "50"]` → `10`; `percentage ["12.5", "80"]` → `10`.
- [x] No changes in `platform/handler` or `platform/server` besides docs/tests.

## Definition of done

`make check` green · ledger updated · commit `Completed: 'T04: Advanced operations'`.
