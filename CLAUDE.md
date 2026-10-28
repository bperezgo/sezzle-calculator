# Project instructions live in AGENTS.md (shared across AI coding tools).

@AGENTS.md

## Claude Code specifics

- A Stop hook (`.claude/hooks/tests-green-gate.sh`) blocks ending a turn while implementation files changed in `backend/` or `frontend/src/` and that side's build/tests are red. If genuinely blocked, say `BLOCKED: <reason>` and hand back to the developer.
- `frontend/CLAUDE.md` loads the Next.js agent rules when working in `frontend/`.
