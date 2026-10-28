# T10 — Final review and submission

- **Area:** repo
- **Depends on:** T09
- **Estimate:** 15 min
- **Status:** ⬜ Todo

## Goal

Ship a clean, reviewable repository and share it.

## Checklist

- [ ] `make check` green locally; CI green on `main`.
- [ ] Self-review: `/code-review` (and `/security-review`) on the full diff; address or consciously defer findings (note deferrals in the README "Assumptions / next steps").
- [ ] Every ticket's acceptance criteria are ticked, each with a test behind it.
- [ ] No secrets in history (gitleaks job green), no `.env` committed, no stray coverage files.
- [ ] `docs/PROMPTS.md` complete, every entry has its *Review* filled.
- [ ] Commit history readable: one `Completed: 'Tnn: …'` per ticket.
- [ ] Push to GitHub (public or shared with the reviewers), verify the README renders, share the link with Sezzle together with `docs/PROMPTS.md`.
