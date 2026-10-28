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
  - govulncheck (second run): `quic-go` (GO-2026-5676, via Gin) → v0.64.1.
  - gitleaks: failed only because the push range starts at the root commit (no parent); passes on later pushes.
- **Review:** _to be filled by the developer._

## #4 — T01 decisions: rounding and input limits

- **Date:** 2026-10-27
- **Ticket:** T01
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > Agree with both, add that to T01

- **Outcome:** T01 decisions 2 and 3 marked accepted: `resultScale = 16` via `DivRound` (global `DivisionPrecision` untouched, half-away-from-zero rounding); operand limits `maxSignificantDigits = 50`, `maxExponent = 1000` → `operand_too_large`. Added acceptance criteria for rounding (`2/3` → `0.6666666666666667`) and both limit boundaries.
- **Review:** _to be filled by the developer._

## #5 — Implement T01

- **Date:** 2026-10-27
- **Ticket:** T01
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > Let's implement @docs/tickets/T01-calculator-domain.md

- **Outcome:** Added `shopspring/decimal` v1.6.0. Tests written first (one table row per acceptance criterion), then:
  - `internal/domain/error`: `*Error{Kind, Code, Message, Fields}`, five sentinels with contract codes/messages as constants, `errors.Is` by code, `WithFields` returns a copy.
  - `internal/domain/calculator`: `Operation` (parsed from a string, carries `Arity` and its evaluator, so T04 just adds registry entries), `Operand` (from string only; ≤ 50 significant digits, scientific exponent within ±1000), `Calculation` via `NewCalculation` (arity check) → `Evaluate()`. Division uses `DivRound(…, 16)`.
  - Coverage 100% on both domain packages; `make check` green.
- **Review:** _to be filled by the developer._

## #6 — Implement T02

- **Date:** 2026-10-27
- **Ticket:** T02
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > Let's start with @docs/tickets/T02-calculation-use-case.md

- **Outcome:** Decision 1 accepted as recommended. Tests written first, then:
  - `internal/domain/dto`: `CalculationRequest`, `CalculationResult`, `OperationInfo` (stdlib-only, decimals as canonical strings).
  - `internal/domain/ports.Calculator`: `Calculate(ctx, req)` and `Operations(ctx)`.
  - `internal/domain/service.CalculatorService`: checks `ctx.Err()` first, maps DTO → value objects → result DTO, echoes canonical operands (`"2.50"` → `"2.5"`), returns domain errors untouched.
  - Domain registry is now an ordered slice with symbols (`+ − × ÷`) plus `calculator.Operations()`.
  - `ports/mocks/calculator.go` generated by `make generate-mocks`; `go mod tidy` added `stretchr/objx`, which testify's mock package needs.
  - `docs/ARCHITECTURE.md` gains a "Request flow" section, including a note that persistence would add a driven port.
  - Coverage is 100% on `domain/service`; `make check` green.
- **Review:** _to be filled by the developer._

## #7 — Implement T03

- **Date:** 2026-10-27
- **Ticket:** T03
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > Let's implement it @docs/tickets/T03-rest-api.md

- **Outcome:** Decisions 4 (result always a string; operands accept strings or JSON numbers, read from raw text) and 5 (`GET /api/v1/operations`) accepted as recommended. Tests written first (handler suite on the mockery mock, middleware unit tests, router end-to-end on the real service), then:
  - `platform/handler/calculation_handler.go`: strict JSON decode (one value, required `operation`/`operands`, operands string-or-number via a custom `UnmarshalJSON`) → `ports.Calculator` → JSON; `Operations` handler.
  - `platform/handler/error_response.go`: the single domain-error → HTTP mapper (400 / 422 / generic 500 with the cause recorded for the log) plus transport codes `invalid_request`, `payload_too_large`, `not_found`, `method_not_allowed`, `internal_error`.
  - `platform/middleware`: request ID (echo if safe, else generated), `slog` access log, JSON panic recovery, 1 MiB body limit.
  - Router: `POST /api/v1/calculations`, `GET /api/v1/operations`, JSON 404/405; wiring in `cmd/main.go`.
  - `docs/api/calculator.md` with curl examples, all verified against `make up`; `docs/ARCHITECTURE.md` request flow updated.
  - Coverage: handler 92.5%, middleware 96.9%, server 100%, domain 100%; `make check` green.
- **Review:** _to be filled by the developer._

## #8 — Drop the Calculator port

- **Date:** 2026-10-27
- **Ticket:** T03 (supersedes part of T02)
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > I don't need a port for the service, you can call it directly from the handler, adding that interface of Calculator increases complexity

- **Outcome:** Removed `internal/domain/ports` (interface + generated mock), `.mockery.yaml`, the `generate-mocks`/`regenerate-mocks` Make targets and the `stretchr/objx` dependency. `NewCalculatorService()` returns `*CalculatorService`; `CalculationHandler` takes it directly. Handler tests rewritten against the real service (domain errors from real inputs, non-domain 500 via a cancelled context); the unknown-`Kind` mapper branch has a small internal test. Updated `AGENTS.md` (concrete types, no interfaces until a second implementation or external dependency), `docs/ARCHITECTURE.md`, the T03 criterion, a "Superseded" note in T02 and the tickets README. `.claude/settings.json` still lists the removed mock targets/path (edit blocked by the auto-mode classifier; harmless). Coverage unchanged (handler 92.5%); `make check` green.
- **Review:** _to be filled by the developer._

## #9 — Acceptance tests in backend/tests

- **Date:** 2026-10-27
- **Ticket:** T03
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > calculation_handler_test.go is the same as an acceptance test, let's create those in a different folder, could be at ./backend/tests, right now we don't have other kind of ports, like repositories, but just for clarifying the responsibilities

- **Outcome:** New `backend/tests/` acceptance suite (black-box HTTP through `server.NewRouter` with the real service, logs captured): `calculations_test.go`, `http_contract_test.go`, shared `app_test.go`. It merges `handler/calculation_handler_test.go` and the end-to-end `server/router_test.go`, both removed; every assertion kept, duplicates collapsed. The unexpected-error case now asserts the cause reaches the access log. Unit tests stay next to the code (middleware, `error_response_test.go`, domain). `make test-race` uses `-coverpkg=./internal/...` so acceptance coverage counts per package (total 100%). `AGENTS.md`, `docs/ARCHITECTURE.md` testing table and the T03 criterion updated; `make check` green.
- **Review:** _to be filled by the developer._

## #10 — Advanced operations

- **Date:** 2026-10-27
- **Ticket:** T04
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > let's continue with @docs/tickets/T04-advanced-operations.md

- **Outcome:** Decision 3 (`0 ^ 0` → `undefined_power`) accepted as recommended; added decision 4 (size guard: `|exponent|` ≤ 1000, power results ≤ 1000 significant digits and exponent within ±1000). Tests first: `backend/tests/advanced_operations_test.go` (one row per criterion, plus a timing check for huge powers), domain rows in `calculation_test.go` (boundaries: `10^1000` / `10^-1000` accepted, `11^1000`, `0.001^500`, `1.1^961` (1001 digits) rejected; `1.1^960` (1000 digits) accepted), and contract rows for the new errors. Then, all in `domain/`:
  - `power` (`^`) via `PowWithPrecision(exp, 16)`: undefined cases are classified before the call → `undefined_power`; `powerMayFit` rejects out-of-range results from a log10 estimate and a digit lower bound before any big-number work, and `resultFits` checks exactly afterwards → `result_too_large`.
  - `sqrt` (`√`, the first unary operation) via `SqrtRound(16)`; negatives → `negative_square_root`.
  - `percentage` (`%`) = `a × b` shifted two places, exact.
  - New `KindMath` errors (422 through the existing mapper). No changes to `platform/handler` or `platform/server`. The existing operation-list tests (domain, service, `GET /api/v1/operations`) were extended with the three new entries.
  - `docs/api/calculator.md`: operations table, size limits, new error codes; curl examples checked against a running server. Worst-case inputs run in < 0.5 ms; coverage 100%; `make check` green.
- **Review:** _to be filled by the developer._

## #11 — Frontend API client and Server Action

- **Date:** 2026-10-27
- **Ticket:** T05
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > Let's continue with @docs/tickets/T05-frontend-api-client.md, and go with the recommended things

- **Outcome:** Recommended decision taken: no validation library (small typed parser). Read the Next 16 Server Actions, forms, instrumentation and `server-only` docs first. Tests first (63 frontend tests total), then:
  - `src/lib/api/types.ts`: `Operation` (+ arity), `CalculationRequest/Response` with string operands/result, `ApiErrorCode` union (every API code + client-side `backend_unavailable`), `CalculateResult` discriminated union.
  - `src/lib/api/client.ts` (`server-only`): `calculate(req, { timeoutMs })` — POST with `cache: "no-store"` and `AbortSignal.timeout` (5 s); network error, timeout, 5xx and any body that breaks the contract (bad JSON, numeric result, unknown operation/error code) → `backend_unavailable`; never throws. Tests cover success, all 12 non-5xx error codes, malformed responses, network error and a real timeout.
  - `src/lib/env.ts` (`server-only`): `parseBackendUrl` validates http(s) and strips the trailing slash; `src/instrumentation.ts` calls it at server start (checked: the built server logs "BACKEND_URL is not set…" when it's missing).
  - `src/lib/validation.ts`: `parseCalculationForm` (fields `operation`, `a`, `b`; unary ops ignore `b`) returning field error codes for T07. `src/lib/calculator-state.ts`: serialisable `useActionState` state that always echoes the inputs. `src/app/actions.ts`: `calculateAction`.
  - `frontend/.env.example`, un-ignored in `frontend/.gitignore`; `server-only` added as a dependency and aliased to its empty module in Vitest.
  - `pnpm build` with a sentinel `BACKEND_URL`: neither name nor value in `.next/static` (re-check in T06 once the page uses the action). `make check` green.
- **Review:** _to be filled by the developer._

## #12 — Calculator UI

- **Date:** 2026-10-27
- **Ticket:** T06
- **Tool / model:** Claude Code · claude-opus-5-5
- **Prompt:**

  > Let's contineu with @docs/tickets/T06-calculator-ui.md

- **Outcome:** Tests first (component + formatter), then:
  - `src/components/calculator/calculator.tsx` (client leaf): labelled first/second number inputs (`inputMode="decimal"`), `<select>` operation picker (`÷ Divide`, …), submit, and an `<output role="status" aria-live="polite">` result panel showing the expression (`10 ÷ 4 =`) and the grouped result with the exact value in `title`. `useActionState`; pending disables submit ("Calculating…"). `sqrt` hides the second operand and relabels the first "Number". Takes the action as a prop so tests need no module mocks; `page.tsx` passes `calculateAction`.
  - `src/lib/format.ts`: thousands grouping on the string itself. Deviation from the ticket's `Intl.NumberFormat` suggestion, found by a test: Intl formats `10^1000` as `∞` and caps fraction digits at 100. Ticket updated.
  - Bug found during the manual check and reproduced in a test first: React 19 resets the form after an action and doesn't restore a controlled `<select>`, so the picker snapped back to "Add". With JS, submission now goes through `onSubmit` + `startTransition(formAction)`, which skips the reset. Without JS, the `action` attribute still posts to the Server Action.
  - `vitest.setup.ts`: Testing Library `cleanup` after each test (auto-cleanup needs Vitest globals).
  - Verified against the running Go API: no-JS multipart POST renders `10 ÷ 4 = 2.5` with inputs kept; at 320 px there's no horizontal scroll and controls are 44 px; a 1000-digit `2^1000` wraps and scrolls; dark mode follows `prefers-color-scheme`. `BACKEND_URL` is still absent from `.next/static`. Error messages are the raw API text for now; T07 maps them. `make check` green.
- **Review:** _to be filled by the developer._
