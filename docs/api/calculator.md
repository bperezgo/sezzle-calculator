# Calculator API

Base URL (local): `http://localhost:8080` — `make up`, or `cd backend && make run`.

Only the Next.js server calls this API; the browser never does, so there is no CORS.

## Conventions

- Every response, errors included, is `application/json; charset=utf-8`.
- Every response carries `X-Request-ID`. A client-sent `X-Request-ID` (1–128 chars of `A-Z a-z 0-9 . _ -`) is echoed; otherwise one is generated. It also appears in the access log.
- **Numbers are exact decimals.** Results are always JSON **strings** (`"0.3"`), because a JSON number would be rounded to float64 by JavaScript clients. Operands may be strings or JSON numbers; numbers are read from their raw text, never through float64.
- Request bodies are limited to 1 MiB.

## `POST /api/v1/calculations`

Request:

| Field | Type | Notes |
|---|---|---|
| `operation` | string | One of the names from `GET /api/v1/operations` |
| `operands` | array of strings or numbers | Length must match the operation's arity |

Response `200`: the operation, the operands in canonical form (`"2.50"` → `"2.5"`), and the result.

```bash
curl -s -X POST localhost:8080/api/v1/calculations \
  -H 'Content-Type: application/json' \
  -d '{"operation":"add","operands":["2","3"]}'
# {"operation":"add","operands":["2","3"],"result":"5"}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"add","operands":[0.1, 0.2]}'
# {"operation":"add","operands":["0.1","0.2"],"result":"0.3"}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"divide","operands":["1","3"]}'
# {"operation":"divide","operands":["1","3"],"result":"0.3333333333333333"}
```

Non-terminating results (division, square root, negative or fractional powers) are rounded half away from zero to 16 decimal places.

| `operation` | Operands | Result |
|---|---|---|
| `add` · `subtract` · `multiply` · `divide` | `[a, b]` | `a + b` · `a − b` · `a × b` · `a ÷ b` |
| `power` | `[base, exponent]` | `base ^ exponent`; exact for positive integer exponents |
| `sqrt` | `[x]` | `√x` |
| `percentage` | `[a, b]` | `a% of b` = `a × b / 100`, exact |

`power` is bounded so arbitrary precision cannot be abused: `|exponent|` ≤ 1000, and the result must have at most 1000 significant digits and an exponent within ±1000 (the same range as operands), otherwise `result_too_large`.

```bash
curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"power","operands":["2","-1"]}'
# {"operation":"power","operands":["2","-1"],"result":"0.5"}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"sqrt","operands":["2"]}'
# {"operation":"sqrt","operands":["2"],"result":"1.414213562373095"}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"percentage","operands":["12.5","80"]}'
# {"operation":"percentage","operands":["12.5","80"],"result":"10"}
```

## `GET /api/v1/operations`

The supported operations, in display order.

```bash
curl -s localhost:8080/api/v1/operations
# {"operations":[{"name":"add","arity":2,"symbol":"+"},{"name":"subtract","arity":2,"symbol":"−"},
#                {"name":"multiply","arity":2,"symbol":"×"},{"name":"divide","arity":2,"symbol":"÷"},
#                {"name":"power","arity":2,"symbol":"^"},{"name":"sqrt","arity":1,"symbol":"√"},
#                {"name":"percentage","arity":2,"symbol":"%"}]}
```

## `GET /health`

```bash
curl -s localhost:8080/health
# {"status":"ok","version":"local"}
```

## Errors

Envelope: `{"error":{"code":"…","message":"…"}}`. `code` is the stable contract to switch on; `message` is human-readable and may change.

| Status | `code` | When |
|---|---|---|
| 400 | `invalid_request` | Body is not JSON, is empty, has trailing data, `operation` is missing or not a string, `operands` is missing or not an array, or an operand is neither a string nor a number (`true`, `null`, …) |
| 400 | `unsupported_operation` | Unknown `operation` |
| 400 | `invalid_operand_count` | Wrong number of operands for the operation |
| 400 | `invalid_operand` | An operand is not a decimal number (`"a"`) |
| 400 | `operand_too_large` | An operand has more than 50 significant digits or an exponent beyond ±1000 |
| 413 | `payload_too_large` | Body over 1 MiB |
| 422 | `division_by_zero` | Well-formed request, mathematically undefined: divisor is zero |
| 422 | `undefined_power` | `0 ^ 0`, `0 ^ y` with `y < 0`, or a negative base with a fractional exponent |
| 422 | `negative_square_root` | `sqrt` of a negative number (no complex results) |
| 422 | `result_too_large` | A `power` result beyond the size limits above (e.g. `10 ^ 100000`) |
| 404 | `not_found` | Unknown route |
| 405 | `method_not_allowed` | Known route, wrong method (`Allow` header lists the right one) |
| 500 | `internal_error` | Unexpected failure; the cause is logged with the request ID, never returned |

```bash
curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"divide","operands":["10","0"]}'
# 422 {"error":{"code":"division_by_zero","message":"division by zero is undefined"}}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"power","operands":["10","100000"]}'
# 422 {"error":{"code":"result_too_large","message":"result exceeds the supported size"}}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"sqrt","operands":["-4"]}'
# 422 {"error":{"code":"negative_square_root","message":"square root of a negative number is undefined"}}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"power","operands":["0","0"]}'
# 422 {"error":{"code":"undefined_power","message":"power is undefined for these operands"}}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"add","operands":["a","1"]}'
# 400 {"error":{"code":"invalid_operand","message":"operand is not a valid decimal number"}}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"add","operands":[true,"1"]}'
# 400 {"error":{"code":"invalid_request","message":"request body must be a JSON object with an operation string and an operands array of numbers or numeric strings"}}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"modulo","operands":["1","2"]}'
# 400 {"error":{"code":"unsupported_operation","message":"operation is not supported"}}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"add","operands":["1"]}'
# 400 {"error":{"code":"invalid_operand_count","message":"wrong number of operands for operation"}}

curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"add","operands":["123456789012345678901234567890123456789012345678901","1"]}'
# 400 {"error":{"code":"operand_too_large","message":"operand exceeds the supported size"}}

curl -s -i localhost:8080/api/v1/calculations
# HTTP/1.1 405 Method Not Allowed
# Allow: POST
# {"error":{"code":"method_not_allowed","message":"method not allowed for this route"}}

curl -s localhost:8080/nope
# 404 {"error":{"code":"not_found","message":"route not found"}}
```

## How the frontend shows errors

The UI never shows `message`; `frontend/src/lib/errors.ts` maps each `code` to its own text (e.g. `division_by_zero` → "You can't divide by zero."). The frontend's API client adds one code of its own, `backend_unavailable`, for network failure, timeout (5 s), any 5xx, or a response that doesn't match this contract.
