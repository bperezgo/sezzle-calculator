# T08 — Full-stack integration (compose + e2e smoke)

- **Area:** repo
- **Depends on:** T04, T07
- **Estimate:** 20 min
- **Status:** ⬜ Todo

## Goal

Prove the two services work together exactly as a reviewer will run them: `docker compose up`.

## Scope

### In
- Verify compose wiring end to end (frontend → `http://backend:8080`), healthchecks, restart policy, sensible resource-free defaults.
- Extend the CI `docker` job smoke test: `POST :8080/api/v1/calculations` returns the expected result, and the frontend page renders.
- Optional (if time): one Playwright test that fills the form and asserts the result, run against the compose stack in CI.

### Out
- Kubernetes, cloud deployment, observability stack.

## Decisions

1. **Playwright** — *Recommendation:* include only if the 2–4 h budget allows; a curl-based smoke test plus component tests already cover the brief.

## Acceptance criteria

- [ ] Fresh clone → `make up` → http://localhost:3000 calculates `2 + 3 = 5` through the backend.
- [ ] CI `docker` job asserts a real calculation through the API.
- [ ] Stopping the backend → the UI shows the `backend_unavailable` message (manual check).

## Definition of done

`make check` green · CI green · ledger updated · commit `Completed: 'T08: Full-stack integration'`.
