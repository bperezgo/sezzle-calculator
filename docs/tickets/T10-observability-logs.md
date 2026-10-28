# T10 — Observability: structured logs in Grafana

- **Area:** repo
- **Depends on:** T09
- **Estimate:** 40 min
- **Status:** ✅ Done

## Goal

Follow one calculation across both services from a local Grafana: every backend call the frontend makes and every request the API serves is a structured JSON log line, correlated by one request ID, shipped by Alloy to Loki. Local only; reuses the Alloy + Grafana setup from the Recurly TIN-validation project, trimmed to logs.

## Scope

### In
- Backend: the access log records the API error `code` of every error response (`error_code`), so failures can be filtered by contract code.
- Frontend: a `server-only` JSON logger (`src/lib/logger.ts`) with the same keys as Go's `slog` JSON handler (`time`, `level`, `msg`), and one log line per backend call in `src/lib/api/client.ts`.
- Frontend → backend correlation: the client sends an `X-Request-ID` it generated; both log lines carry it as `request_id`.
- `deploy/local/`: Alloy config (Docker socket → this compose project's containers → Loki) and Grafana datasource provisioning.
- Compose services `alloy`, `loki`, `grafana` behind the `observability` profile; `make up-obs`; `make down` also stops them.
- README "Local observability" section; ARCHITECTURE deployment note.

### Out
- Metrics and traces (no OTLP, Prometheus or Tempo).
- Provisioned dashboards, alerting, log retention or persistence (telemetry dies with `make down`).
- Anything beyond localhost.

## Decisions

1. **Opt-in profile vs. part of `make up`** — ✅ **profile `observability`** (`make up-obs`). The reviewer quick start and the CI `docker` job stay as fast as today; the telemetry stack is a local extra.
2. **Who creates the request ID** — ✅ **the frontend** (`crypto.randomUUID()` per backend call), sent as `X-Request-ID`. The backend already echoes a valid incoming ID, so no backend change is needed for correlation.
3. **Parsing** — ✅ Alloy promotes only `level` to a Loki label (low cardinality); every other field stays in the JSON line and is extracted at query time with `| json`. Non-JSON lines (Next.js start-up banner) pass through untouched. Docker healthcheck `/health` lines are dropped in Alloy (the app still logs them): one every 5 s buried the real traffic.
4. **What the frontend logs** — ✅ operation, status, outcome, duration and request ID. Operands and results are not logged: they are user input and add nothing to diagnosing the service.

## Acceptance criteria

Backend (`middleware`/`handler` unit tests, `backend/tests` acceptance tests):

- [x] A `POST /api/v1/calculations` that fails with `division_by_zero` produces an access log line with `"status":422` and `"error_code":"division_by_zero"`.
- [x] A successful request's access log line has no `error_code`.
- [x] A panic → 500 access log line has `"error_code":"internal_error"`.

Frontend (`src/lib/logger.test.ts`, `src/lib/api/client.test.ts`):

- [x] `logger.info/error(msg, attrs)` writes one JSON line to stdout with `time` (ISO 8601), `level` (`INFO`/`ERROR`), `msg` and the attrs.
- [x] Every backend call sends an `X-Request-ID` header matching the backend's accepted charset, unique per call.
- [x] Success → one `INFO` `"backend call"` line with `request_id` (same as the header), `operation`, `status: 200`, `outcome: "ok"`, `duration_ms`.
- [x] API error (e.g. 422 `division_by_zero`) → `INFO` line with `outcome: "api_error"`, `error_code: "division_by_zero"`.
- [x] Backend unavailable → `ERROR` line with `outcome: "backend_unavailable"` and a `reason` of `network`, `timeout`, `server_error` (5xx, with `status`) or `invalid_response`.
- [x] No log line contains the operands or the result.

Local stack (manual, recorded in the ledger):

- [x] `make up-obs && make smoke` → Grafana at http://localhost:3001 (no login) shows `{service_name="backend"}` and `{service_name="frontend"}` logs in Explore within ~10 s. (Verified through Grafana's Loki proxy on alternate ports: smoke traffic visible, `level` label set, `error_code` on the 422, `/health` lines dropped; backend stopped → frontend `ERROR` line with `reason: network`.)
- [ ] Clicking a `request_id` in a frontend line opens the backend line for the same request (Loki derived field). (Provisioning verified via the Grafana API; the click in the UI is still to be checked by the developer — tracked in T11.)
- [x] `make up` alone starts only backend + frontend (CI unchanged); `make down` removes everything. (`docker compose config` without the profile lists only backend + frontend; after `make down` no project containers remain.)

## Definition of done

`make check` green · manual Grafana check done · ledger updated · commit `Completed: 'T10: Observability: structured logs in Grafana'`.
