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

Non-terminating division results are rounded half away from zero to 16 decimal places.

## `GET /api/v1/operations`

The supported operations, in display order.

```bash
curl -s localhost:8080/api/v1/operations
# {"operations":[{"name":"add","arity":2,"symbol":"+"},{"name":"subtract","arity":2,"symbol":"−"},
#                {"name":"multiply","arity":2,"symbol":"×"},{"name":"divide","arity":2,"symbol":"÷"}]}
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
| 422 | `division_by_zero` | Well-formed request, mathematically undefined |
| 404 | `not_found` | Unknown route |
| 405 | `method_not_allowed` | Known route, wrong method (`Allow` header lists the right one) |
| 500 | `internal_error` | Unexpected failure; the cause is logged with the request ID, never returned |

```bash
curl -s -X POST localhost:8080/api/v1/calculations \
  -d '{"operation":"divide","operands":["10","0"]}'
# 422 {"error":{"code":"division_by_zero","message":"division by zero is undefined"}}

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

curl -s -i localhost:8080/api/v1/calculations
# HTTP/1.1 405 Method Not Allowed
# Allow: POST
# {"error":{"code":"method_not_allowed","message":"method not allowed for this route"}}
```
