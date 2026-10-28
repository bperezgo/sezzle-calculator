# Tickets

The plan to deliver the brief in `test.md` (budget ~2–4 h of build time). Work them **in order**; each one is small enough for a single `Completed:` commit. **The acceptance criteria are the spec** — there is no separate spec tool.

| ID | Title | Area | Depends on | Est. | Status |
|---|---|---|---|---|---|
| [T00](T00-project-setup.md) | Project setup | repo | — | — | ✅ Done |
| [T01](T01-calculator-domain.md) | Calculator domain: basic operations | backend | T00 | 30 min | ✅ Done |
| [T02](T02-calculation-use-case.md) | Calculation use case and ports | backend | T01 | 20 min | ✅ Done |
| [T03](T03-rest-api.md) | REST API with Gin | backend | T02 | 40 min | ✅ Done |
| [T04](T04-advanced-operations.md) | Advanced operations: power, square root, percentage | backend | T03 | 25 min | ⬜ Todo |
| [T05](T05-frontend-api-client.md) | Frontend API client and Server Action | frontend | T03 | 30 min | ⬜ Todo |
| [T06](T06-calculator-ui.md) | Calculator UI | frontend | T05 | 45 min | ⬜ Todo |
| [T07](T07-error-handling-ux.md) | Validation and error handling UX | frontend | T06 | 20 min | ⬜ Todo |
| [T08](T08-full-stack-integration.md) | Full-stack integration (compose + e2e smoke) | repo | T04, T07 | 20 min | ⬜ Todo |
| [T09](T09-documentation.md) | Documentation and coverage report | docs | T08 | 25 min | ⬜ Todo |
| [T10](T10-submission.md) | Final review and submission | repo | T09 | 15 min | ⬜ Todo |

Status legend: ⬜ Todo · 🟨 In progress · ✅ Done. Update the table when a ticket changes state.

## Workflow per ticket

1. **Read** the ticket. Resolve its *Decisions* first (agree or override the recommendation) and record the outcome in the ticket.
2. **Plan:** review the acceptance criteria together; add or sharpen criteria until each one has a concrete input and an exact expected output (value, error code, HTTP status, UI text).
3. **Test first:** write one test (or table row) per criterion, named after it, and watch it fail.
4. **Implement** until green — backend domain-first (value objects → service → handler → wiring).
5. **Verify:** `make check` green (and `make up` when runtime wiring changed). Tick the criteria.
6. **Ledger:** append the prompts used to `docs/PROMPTS.md`; fill in *Review*.
7. **Commit:** small commits allowed; the last one is `Completed: 'Tnn: <title>'`. Mark the ticket ✅ here.

## Ticket template

```markdown
# Tnn — Title

- **Area:** backend | frontend | repo | docs
- **Depends on:** Tnn
- **Estimate:** NN min
- **Status:** ⬜ Todo

## Goal
## Scope
### In
### Out
## Decisions
## Acceptance criteria
## Definition of done
```
