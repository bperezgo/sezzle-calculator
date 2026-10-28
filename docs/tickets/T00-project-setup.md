# T00 — Project setup

- **Area:** repo
- **Depends on:** —
- **Status:** ✅ Done (ledger #1)

## Goal

A monorepo where both apps build, lint, test and run together, so every later ticket is pure feature work.

## Delivered

- [x] `backend/`: Go module `sezzle.com/calculator`, Gin router with `GET /health`, env config (`PORT`, `GIN_MODE`), graceful shutdown, `-healthcheck` flag for the distroless image.
- [x] `backend/.golangci.yml` (v2; depguard keeps `internal/domain` free of platform/Gin/`net/http`/`encoding/json`), `.mockery.yaml`, `Makefile`, `Dockerfile` (multi-stage, distroless, non-root, pinned digests).
- [x] `frontend/`: Next.js 16.4 App Router, React 19, TypeScript strict, Tailwind v4, ESLint, Prettier (+ Tailwind class sorting), Vitest + Testing Library + coverage, `output: "standalone"`, `Dockerfile`.
- [x] Root `docker-compose.yml`: backend (healthcheck) → frontend with `BACKEND_URL=http://backend:8080`.
- [x] Root `Makefile` (`check`, `cover`, `up`, `down`, `install-hooks`), `lefthook.yml`.
- [x] `.github/workflows/ci.yml`: backend lint · test (race + coverage artifact + job summary) · govulncheck; frontend lint · format · typecheck · test (coverage artifact) · build; docker compose smoke test; gitleaks; hadolint. Actions pinned to commit SHAs. Dependabot for gomod, npm, actions, docker.
- [x] `AGENTS.md`, `CLAUDE.md`, `.claude/settings.json` (permissions, Stop hook that blocks on red tests), `docs/ARCHITECTURE.md`, `README.md` skeleton, `docs/PROMPTS.md`, these tickets.

## Verified

- `make -C backend check` → 0 lint issues, tests green.
- `cd frontend && pnpm check && pnpm build` → green.
- `docker compose up --build --wait` → both containers healthy; `GET :8080/health` → `{"status":"ok","version":"local"}`, `GET :3000/` → 200.
- `actionlint` and `hadolint` clean.

## Follow-ups

- Push to GitHub and confirm the first CI run is green (T10 at the latest; ideally right after T00 is committed).

## Changes after review (ledger #2)

- OpenSpec removed: the tickets' acceptance criteria already are the spec, and each one maps to a test. One source of truth instead of ticket + proposal + spec + tasks for a 2–4 h project.
