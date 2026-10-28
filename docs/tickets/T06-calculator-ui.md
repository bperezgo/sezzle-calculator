# T06 — Calculator UI

- **Area:** frontend
- **Depends on:** T05
- **Estimate:** 45 min
- **Status:** ⬜ Todo

## Goal

An intuitive, responsive calculator that renders on the server and enhances on the client.

## Scope

### In
- `src/app/page.tsx` (Server Component): layout, heading, renders the calculator.
- `src/components/calculator/` (Client Component leaf): operands, operation picker, submit, result panel. Uses `useActionState(calculateAction)`; pending state disables submit.
- Unary operations (sqrt) hide/disable the second operand.
- Result display in `src/lib/format.ts`: the backend's exact decimal string, grouped for readability (`Intl.NumberFormat` accepts decimal strings without converting to float). Long results (`1/3`) wrap or truncate with the full value available (e.g. `title`/copy). Never round-trip through `number`.
- Responsive: single column on phones (≥ 320 px), comfortable tap targets (≥ 44 px), light/dark via `prefers-color-scheme`.
- Accessibility: every input labelled, result and errors in an `aria-live="polite"` region, full keyboard use, visible focus.

### Out
- Calculation history, expression parsing (`2+3*4`), keyboard shortcuts beyond normal form behaviour.

## Decisions

1. **Styling** — ✅ **Tailwind CSS v4** (already scaffolded). It's zero-runtime (works cleanly with Server Components, unlike CSS-in-JS), needs no config file in v4, you already know it, and the UI is small. Alternatives considered: **CSS Modules** (zero dependencies, more files, fine but slower to iterate); **shadcn/ui** (copy-in Radix components — accessible but overkill for one form, adds several deps).
2. **Interaction model** — ✅ **form-based** to start (developer has doubts — revisit after the first working version; the hybrid with operation buttons is the cheapest pivot) (`[number] [operation ▾] [number] [=]`). It maps 1:1 to the API, works without JavaScript (progressive enhancement via Server Action), and is easy to validate. Alternative: a **keypad** (pocket-calculator buttons) — more familiar, but needs client state for the expression and doesn't work before hydration. A hybrid (operation buttons instead of a `<select>`) is a cheap middle ground.

## Acceptance criteria

- [ ] Entering `10`, `÷`, `4` and submitting shows `2.5`.
- [ ] Works with JavaScript disabled (form posts to the Server Action, page re-renders with the result).
- [ ] Usable at 320 px width with no horizontal scroll.
- [ ] Component tests: renders all operations, unary op hides second operand, shows result from action state, submit disabled while pending.
- [ ] No ESLint/a11y warnings.

## Definition of done

`make check` green · manual check in `pnpm dev` (desktop + mobile viewport) · ledger updated · commit `Completed: 'T06: Calculator UI'`.
