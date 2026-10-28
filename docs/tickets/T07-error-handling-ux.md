# T07 — Validation and error handling UX

- **Area:** frontend
- **Depends on:** T06
- **Estimate:** 20 min
- **Status:** ✅ Done

## Goal

Every way a calculation can fail produces a clear, specific message next to the right field, and the app never shows a raw error or crashes.

## Scope

### In
- Client-side validation before submit: required, numeric (accept `-`, `.`, `1e3`), finite; inline field errors (`aria-invalid`, `aria-describedby`).
- Mapping of every API error `code` to a user-facing message in one place (`src/lib/errors.ts`), e.g. `division_by_zero` → "You can't divide by zero.", `backend_unavailable` → "The calculator service is unavailable. Try again."
- Server-side re-validation in the action (already in T05) surfaces field errors in the same shape.
- `src/app/error.tsx` error boundary and a friendly `not-found.tsx`.
- Preserve the user's inputs after an error.

### Out
- Toasts/notifications libraries, retry with backoff.

## Acceptance criteria

- [x] Empty operand → "Enter a number" without calling the server.
- [x] `abc` → "Enter a valid number".
- [x] `10 ÷ 0` → backend 422 → "You can't divide by zero." and inputs kept.
- [x] `sqrt -4` → specific message ("You can't take the square root of a negative number.").
- [x] Backend down → `backend_unavailable` message, no crash.
- [x] Tests for each row above. (`src/components/calculator/calculator.test.tsx` › "Calculator error handling", `src/lib/errors.test.ts`, `src/app/error.test.tsx`, `src/app/not-found.test.tsx`)

## Definition of done

`make check` green · ledger updated · commit `Completed: 'T07: Error handling UX'`.
