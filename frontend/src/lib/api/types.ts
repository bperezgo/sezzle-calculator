// The Go API contract (docs/api/calculator.md). Operands and results are exact decimal
// strings; never convert them to `number`.

export const OPERATIONS = [
  "add",
  "subtract",
  "multiply",
  "divide",
  "power",
  "sqrt",
  "percentage",
] as const;

export type Operation = (typeof OPERATIONS)[number];

export const OPERATION_ARITY: Record<Operation, 1 | 2> = {
  add: 2,
  subtract: 2,
  multiply: 2,
  divide: 2,
  power: 2,
  sqrt: 1,
  percentage: 2,
};

export type CalculationRequest = {
  operation: Operation;
  operands: string[];
};

export type CalculationResponse = {
  operation: Operation;
  operands: string[];
  result: string;
};

export const API_ERROR_CODES = [
  "invalid_request",
  "unsupported_operation",
  "invalid_operand_count",
  "invalid_operand",
  "operand_too_large",
  "payload_too_large",
  "division_by_zero",
  "undefined_power",
  "negative_square_root",
  "result_too_large",
  "not_found",
  "method_not_allowed",
  "internal_error",
] as const;

// `backend_unavailable` is produced by the client, never by the API: network failure,
// timeout, 5xx or a response that doesn't match the contract.
export type ApiErrorCode =
  (typeof API_ERROR_CODES)[number] | "backend_unavailable";

export type ApiError = {
  code: ApiErrorCode;
  message: string;
};

export type CalculateResult =
  { ok: true; result: CalculationResponse } | { ok: false; error: ApiError };
