# T08 — Full-stack integration (compose + e2e smoke)

- **Area:** repo
- **Depends on:** T04, T07
- **Estimate:** 20 min
- **Status:** ✅ Done

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

1. **Playwright** — ✅ **skipped**, as recommended: `scripts/smoke.sh` already drives the real form (no-JS multipart post of the server-rendered Server Action fields) through the backend, and component tests cover the JS path.

## Acceptance criteria

- [x] Fresh clone → `make up` → http://localhost:3000 calculates `2 + 3 = 5` through the backend. (`git clone` → `make up` → `make smoke`: the frontend form post renders `2 + 3 = 5`)
- [x] CI `docker` job asserts a real calculation through the API. (`scripts/smoke.sh`: `2 + 3` → `"5"`, `10 ÷ 0` → 422 `division_by_zero`, and the frontend form → `2 + 3 = 5`)
- [x] Stopping the backend → the UI shows the `backend_unavailable` message (manual check). (`docker compose stop backend` → form post returns 200 with "The calculator service is unavailable. Try again.")

## Definition of done

`make check` green · CI green · ledger updated · commit `Completed: 'T08: Full-stack integration'`.
