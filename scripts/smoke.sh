#!/usr/bin/env bash
# Smoke test against a running stack (`make up`): a real calculation through the API, and
# through the frontend's no-JS form post (page → Server Action → Go API → rendered result).
set -euo pipefail

BACKEND="${BACKEND:-http://localhost:${BACKEND_PORT:-8080}}"
FRONTEND="${FRONTEND:-http://localhost:${FRONTEND_PORT:-3000}}"

fail() { echo "FAIL: $*" >&2; exit 1; }

curl --fail --silent --show-error "$BACKEND/health" >/dev/null
echo "ok   backend health"

body=$(curl --fail --silent --show-error -X POST "$BACKEND/api/v1/calculations" \
  -H 'Content-Type: application/json' -d '{"operation":"add","operands":["2","3"]}')
[[ "$body" == '{"operation":"add","operands":["2","3"],"result":"5"}' ]] || fail "API 2 + 3: $body"
echo "ok   API 2 + 3 = 5"

body=$(curl --silent --show-error -w '\n%{http_code}' -X POST "$BACKEND/api/v1/calculations" \
  -H 'Content-Type: application/json' -d '{"operation":"divide","operands":["10","0"]}')
[[ "$body" == *'"code":"division_by_zero"'*$'\n'422 ]] || fail "API 10 / 0: $body"
echo "ok   API 10 / 0 → 422 division_by_zero"

page=$(curl --fail --silent --show-error "$FRONTEND/")
[[ "$page" == *'<h1'*'Calculator</h1>'* ]] || fail "frontend page did not render the calculator"
echo "ok   frontend renders"

# Re-post the server-rendered form's hidden Server Action fields with the operands filled in.
form=()
while IFS= read -r input; do
  name=$(sed -E 's/.*name="([^"]*)".*/\1/' <<<"$input")
  value=""
  [[ "$input" == *'value="'* ]] && value=$(sed -E 's/.*value="([^"]*)".*/\1/; s/&quot;/"/g; s/&amp;/\&/g' <<<"$input")
  form+=(--form-string "$name=$value")
done < <(grep -oE '<input type="hidden"[^>]*>' <<<"$page")
[[ ${#form[@]} -gt 0 ]] || fail "no Server Action fields in the page"

result=$(curl --fail --silent --show-error "$FRONTEND/" "${form[@]}" \
  --form-string a=2 --form-string operation=add --form-string b=3 | sed 's/<!-- -->//g')
[[ "$result" == *'2 + 3 ='*'>5</span>'* ]] || fail "frontend form post did not render 2 + 3 = 5"
echo "ok   frontend form 2 + 3 = 5 (through the backend)"
