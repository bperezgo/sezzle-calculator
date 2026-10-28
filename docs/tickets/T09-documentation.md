# T09 — Documentation and coverage report

- **Area:** docs
- **Depends on:** T08
- **Estimate:** 25 min
- **Status:** ⬜ Todo

## Goal

Everything the brief's *Deliverables* and *Make sure your README includes* sections ask for.

## Scope

### In
- `README.md`: setup instructions; how to run frontend and backend (compose and locally); API examples (curl for every operation + error cases, with real responses); design decisions and assumptions (hexagonal backend, server-first Next.js, single calculations endpoint, float64, error envelope, percentage semantics, 422 vs 400, why no CORS); testing and coverage; AI usage → `docs/PROMPTS.md`.
- Coverage report: `make cover` produces `backend/coverage.html` and `frontend/coverage/index.html`; README states the current numbers; CI uploads both as artifacts (already in place since T00).
- Finalise `docs/ARCHITECTURE.md` and `docs/api/calculator.md`.
- Optional: a small architecture diagram (Mermaid in the README).

### Out
- Hosted docs site, OpenAPI generator (a hand-written `docs/api/openapi.yaml` is a nice-to-have if time allows).

## Acceptance criteria

- [ ] A reviewer can go from clone to running app using only the README.
- [ ] Every curl example in the README was run and its output pasted.
- [ ] Coverage numbers in the README match a fresh `make cover`.

## Definition of done

Ledger updated · commit `Completed: 'T09: Documentation'`.
