import "server-only";

import { backendUrl } from "@/lib/env";

import {
  API_ERROR_CODES,
  OPERATIONS,
  type ApiError,
  type ApiErrorCode,
  type CalculateResult,
  type CalculationRequest,
  type CalculationResponse,
} from "./types";

const DEFAULT_TIMEOUT_MS = 5000;

export const BACKEND_UNAVAILABLE_MESSAGE =
  "the calculator service is unavailable";

const unavailable: CalculateResult = {
  ok: false,
  error: { code: "backend_unavailable", message: BACKEND_UNAVAILABLE_MESSAGE },
};

type CalculateOptions = { timeoutMs?: number };

export async function calculate(
  request: CalculationRequest,
  { timeoutMs = DEFAULT_TIMEOUT_MS }: CalculateOptions = {},
): Promise<CalculateResult> {
  let response: Response;
  let body: unknown;
  try {
    response = await fetch(`${backendUrl()}/api/v1/calculations`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
      cache: "no-store",
      signal: AbortSignal.timeout(timeoutMs),
    });
    if (response.status >= 500) return unavailable;
    body = await response.json();
  } catch {
    // Network failure, timeout (TimeoutError) or a body that isn't JSON.
    return unavailable;
  }

  if (response.ok) {
    return isCalculationResponse(body)
      ? { ok: true, result: body }
      : unavailable;
  }
  const error = apiError(body);
  return error ? { ok: false, error } : unavailable;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isCalculationResponse(body: unknown): body is CalculationResponse {
  return (
    isRecord(body) &&
    typeof body.operation === "string" &&
    (OPERATIONS as readonly string[]).includes(body.operation) &&
    Array.isArray(body.operands) &&
    body.operands.every((operand) => typeof operand === "string") &&
    typeof body.result === "string"
  );
}

function apiError(body: unknown): ApiError | undefined {
  if (!isRecord(body) || !isRecord(body.error)) return undefined;
  const { code, message } = body.error;
  if (typeof code !== "string" || typeof message !== "string") return undefined;
  if (!(API_ERROR_CODES as readonly string[]).includes(code)) return undefined;
  return { code: code as ApiErrorCode, message };
}
