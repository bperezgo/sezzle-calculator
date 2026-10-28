#!/usr/bin/env bash
# Stop hook: once implementation files changed in backend/ (.go) or frontend/src (.ts/.tsx), refuse to
# end the turn while that side's build or tests are red. Fails open on environment problems and
# releases after 5 consecutive blocks so a genuinely stuck suite can be handed back to the developer.
set -uo pipefail

cd "${CLAUDE_PROJECT_DIR:-.}" 2>/dev/null || exit 0
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || exit 0

count_file=".claude/.gate-block-count"
changed() { git status --porcelain --untracked-files=all -- "$@" 2>/dev/null | awk '{ print $NF }'; }

backend_changed="$(changed 'backend/*.go' | grep -Ev '_test\.go$' || true)"
frontend_changed="$(changed 'frontend/src/*.ts' 'frontend/src/*.tsx' | grep -Ev '\.test\.tsx?$' || true)"

if [ -z "$backend_changed" ] && [ -z "$frontend_changed" ]; then
  rm -f "$count_file"
  exit 0
fi

log="$(mktemp -t calc-gate.XXXXXX)"
trap 'rm -f "$log"' EXIT
ok=1

if [ -n "$backend_changed" ] && command -v go >/dev/null 2>&1; then
  (cd backend && go build ./... && go test ./... -count=1) >>"$log" 2>&1 || ok=0
fi
if [ -n "$frontend_changed" ] && command -v pnpm >/dev/null 2>&1 && [ -d frontend/node_modules ]; then
  (cd frontend && pnpm typecheck && pnpm test) >>"$log" 2>&1 || ok=0
fi

if [ "$ok" -eq 1 ]; then
  rm -f "$count_file"
  exit 0
fi

n=$(( $(cat "$count_file" 2>/dev/null || echo 0) + 1 ))
echo "$n" > "$count_file"
if [ "$n" -ge 5 ]; then
  rm -f "$count_file"
  echo "Suite still red after ${n} attempts — releasing. Report BLOCKED: <reason> to the developer." >&2
  exit 0
fi

{
  echo "BLOCKED (${n}/5): implementation files changed but build/tests are red. Fix before finishing."
  echo
  grep -E -- '--- FAIL|FAIL|panic:|\.go:[0-9]+|error TS|✗|×' "$log" | head -n 40
} >&2
exit 2
