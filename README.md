# Sezzle Calculator

Full-stack calculator: a **Go** REST microservice (Gin, hexagonal architecture) and a **Next.js** (React + TypeScript) frontend that consumes it server-side. Interview brief: [`test.md`](test.md).

> Work in progress — the plan lives in [`docs/tickets/`](docs/tickets/README.md). This README is completed in T09 (setup, API examples, design decisions).

## Quick start

Requirements: Docker. For local development also Go 1.26+, Node 24+, pnpm 11, golangci-lint v2.

```bash
make up        # backend on :8080, frontend on :3000 (docker compose)
make down
```

Local development:

```bash
make install           # pnpm install + go mod download
make install-hooks     # lefthook git hooks
cd backend && make run       # :8080
cd frontend && BACKEND_URL=http://localhost:8080 pnpm dev   # :3000
make check             # backend lint+tests, frontend lint+format+typecheck+tests
make cover             # backend/coverage.html, frontend/coverage/index.html
```

## Repository layout

```
backend/      Go API — cmd/main.go, internal/domain (pure), internal/platform (adapters), Dockerfile
frontend/     Next.js app — src/app, src/components, src/lib, Dockerfile
docs/         ARCHITECTURE.md, tickets/, PROMPTS.md (AI prompt audit ledger)
.github/      CI (lint, tests + coverage, govulncheck, docker smoke test, gitleaks, hadolint), Dependabot
```

## AI usage

Every prompt used to build this project is recorded verbatim in [`docs/PROMPTS.md`](docs/PROMPTS.md).
