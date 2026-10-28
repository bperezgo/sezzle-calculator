# AI Prompt Ledger

Audit trail of every prompt used to build this project, as requested by the brief ("Share any prompts that you used in your work").

**Rules**

- Append-only. One entry per prompt, numbered in order. Never edit or delete an earlier entry — corrections go in a new entry.
- The prompt is copied **verbatim** (typos included). Only secrets are redacted, as `[REDACTED]`.
- *Outcome* is what the AI produced; *Review* is what the developer checked, changed or rejected. Review is filled by the developer.
- Slash commands (`/opsx:propose T01`, …) are logged too, with the arguments used.

**Entry template**

```markdown
## #N — <short title>

- **Date:** YYYY-MM-DD
- **Ticket:** Tnn
- **Tool / model:** Claude Code · <model id>
- **Prompt:**

  > verbatim prompt

- **Outcome:** what was produced (files, decisions, commands run).
- **Review:** what the developer verified / changed / rejected.
```

---

## #1 — Project bootstrap

- **Date:** 2026-10-27
- **Ticket:** T00
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > I need to set up a new project, it is for an interview for a company called Sezzle, I will give you the basic details about the test, there are in @test.md .
  > I implemented another test at ../../template/, there I used hexagonal architecture, domain driven design, we didn't use gin gonic, but I want that in this project, we don't need repositories, but we can use ports and adapters.
  > I want a monorepo approach, I need the github ci, linters, tests, dockerfile (inside the backend), we need two folders, one for the backend, another for the frontend.
  > I will need a docker compose file, where we can run both projects at the same time.
  > For the frontend we will use react js, and typescript, and pnpm as well, I am used to tailwind, but let's discuss the best option, as far as I can see, we don't need a lot of interaction, so I want the project server side from the beginning, so we can use nextjs.
  > We need to use openspec, so install the required tools, you can use, as reference, the ones at ../template and ../laguna-escondida-backend.
  > Another they are asking is a way to review the prompts I will use, so let's create a markdown file, like an audit ledger, beginning with this prompt.
  > I want to work every step, so, let's create, in a docs folder, a set of tickets that we can follow to complete this project

- **Outcome:**
  - Monorepo skeleton: `backend/` (Go 1.26, Gin, `/health`, config, graceful shutdown, distroless `Dockerfile` with self-healthcheck, golangci-lint v2 with depguard hexagonal rule, mockery config, Makefile) and `frontend/` (Next.js 16.4 App Router, React 19, TypeScript, Tailwind v4, Vitest + Testing Library, Prettier, ESLint, standalone `Dockerfile`).
  - Root `docker-compose.yml` (backend :8080 healthy → frontend :3000, `BACKEND_URL` server-only), root `Makefile`, `lefthook.yml`.
  - GitHub Actions CI (backend lint/test+coverage/govulncheck, frontend lint/format/typecheck/test+coverage/build, OpenSpec validate, docker compose smoke test, gitleaks, hadolint), Dependabot, PR template. Actions pinned to SHAs.
  - OpenSpec initialised for Claude Code (`/opsx:*` commands + skills), `openspec/config.yaml` with project context and rules.
  - `AGENTS.md` / `CLAUDE.md`, `.claude/settings.json` (permissions + Stop hook gating red tests), `docs/ARCHITECTURE.md`, `README.md` skeleton, this ledger, and tickets T00–T10 in `docs/tickets/`.
  - Styling discussion: recommended keeping Tailwind v4 (see T06 decision); confirmation pending.
- **Review:** _to be filled by the developer._

## #2 — Ticket decisions and dropping OpenSpec

- **Date:** 2026-10-27
- **Ticket:** T00 (planning for T01, T03, T04, T06)
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > for T01 we can use Decimals https://github.com/shopspring/decimal, it is battle tested, and can be more accurate.
  > For T03, I like the approach
  > For T04, use the recommended approach
  > For T06, I have my doubts, but let's start with this
  >
  > Seeing all the specification in these ticket, I don't think we will need to use openspec, as far as we can have acceptance criteria, so we can create tests, to match the ticket we can do it ok, what do you think?

- **Outcome:**
  - Agreed to drop OpenSpec: tickets with concrete acceptance criteria are the single spec; each criterion maps to a named test. Removed `openspec/`, the `/opsx:*` commands and skills, the CI job, Makefile/lefthook targets and permissions; rewrote the per-ticket workflow (`docs/tickets/README.md`) and the "Tickets are the spec" rule in `AGENTS.md`.
  - T01 rewritten around `shopspring/decimal` v1.6.0 (checked its API: `DivRound`, `SqrtRound`, `PowWithPrecision`). Added decisions: 16-place rounding via a domain constant, operand size limits (`operand_too_large`).
  - T02: decimals cross layers as canonical strings in DTOs. T03: decisions marked accepted; new recommendation — result is a JSON string, operands accept strings or JSON numbers. T04: percentage = a% of b, sqrt of negative = error, power via `PowWithPrecision`, `result_too_large` guard. T05/T06: values stay strings in the frontend; Tailwind + form-based UI accepted, to revisit after the first version.
- **Review:** _to be filled by the developer._

## #3 — Commit and push T00

- **Date:** 2026-10-27
- **Ticket:** T00
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > Let's commit it, github repo was already created, so we only need to push to github

- **Outcome:** Committed the project setup as `Completed: 'T00: Project setup'` on `main` and pushed to `origin` (github.com/bperezgo/sezzle-calculator). The first CI run failed and surfaced real issues, fixed in a follow-up commit:
  - govulncheck: reachable stdlib `crypto/x509` vuln (GO-2026-6809, fixed in Go 1.26.10) and `golang.org/x/net` HPACK race (GO-2026-6617) → `go 1.26.10`, builder image `golang:1.26.10-alpine`, `x/net` v0.62.0.
  - golangci-lint v2.6.2 binary is built with Go 1.25 and refuses a Go 1.26 module → CI uses v2.15.0.
  - Docker build: empty `frontend/public/` was not tracked by git, so the clean checkout lacked it → `public/.gitkeep`.
  - gitleaks: failed only because the push range starts at the root commit (no parent); passes on later pushes.
- **Review:** _to be filled by the developer._
