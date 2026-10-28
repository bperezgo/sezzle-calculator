import "server-only";

import { backendUrl } from "@/lib/env";
import { logger, type LogAttrs } from "@/lib/logger";

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

// The Go API echoes a valid incoming ID, so one ID ties the frontend and backend log lines together.
const REQUEST_ID_HEADER = "X-Request-ID";

export const BACKEND_UNAVAILABLE_MESSAGE =
  "the calculator service is unavailable";

const unavailable: CalculateResult = {
  ok: false,
  error: { code: "backend_unavailable", message: BACKEND_UNAVAILABLE_MESSAGE },
};

type CalculateOptions = { timeoutMs?: number };

type UnavailableReason =
  "network" | "timeout" | "server_error" | "invalid_response";

// What one backend call produced, plus the attrs that describe it in the log.
type CallOutcome = { result: CalculateResult; log: LogAttrs };

// Operands and results are user input; they are never logged.
export async function calculate(
  request: CalculationRequest,
  { timeoutMs = DEFAULT_TIMEOUT_MS }: CalculateOptions = {},
): Promise<CalculateResult> {
  const requestId = crypto.randomUUID();
  const start = performance.now();
  const { result, log } = await callBackend(request, requestId, timeoutMs);

  const attrs: LogAttrs = {
    request_id: requestId,
    operation: request.operation,
    ...log,
    duration_ms: Math.round((performance.now() - start) * 10) / 10,
  };
  if (!result.ok && result.error.code === "backend_unavailable") {
    logger.error("backend call", attrs);
  } else {
    logger.info("backend call", attrs);
  }
  return result;
}

async function callBackend(
  request: CalculationRequest,
  requestId: string,
  timeoutMs: number,
): Promise<CallOutcome> {
  let response: Response;
  try {
    response = await fetch(`${backendUrl()}/api/v1/calculations`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        [REQUEST_ID_HEADER]: requestId,
      },
      body: JSON.stringify(request),
      cache: "no-store",
      signal: AbortSignal.timeout(timeoutMs),
    });
  } catch (error) {
    return unavailableOutcome(isTimeout(error) ? "timeout" : "network");
  }

  const { status } = response;
  if (status >= 500) return unavailableOutcome("server_error", status);

  let body: unknown;
  try {
    body = await response.json();
  } catch (error) {
    return unavailableOutcome(
      isTimeout(error) ? "timeout" : "invalid_response",
      status,
    );
  }

  if (response.ok) {
    return isCalculationResponse(body)
      ? { result: { ok: true, result: body }, log: { status, outcome: "ok" } }
      : unavailableOutcome("invalid_response", status);
  }
  const error = apiError(body);
  return error
    ? {
        result: { ok: false, error },
        log: { status, outcome: "api_error", error_code: error.code },
      }
    : unavailableOutcome("invalid_response", status);
}

function unavailableOutcome(
  reason: UnavailableReason,
  status?: number,
): CallOutcome {
  const log: LogAttrs = { outcome: "backend_unavailable", reason };
  if (status !== undefined) log.status = status;
  return { result: unavailable, log };
}

// AbortSignal.timeout rejects with a DOMException named TimeoutError, which isn't always an Error subclass.
function isTimeout(error: unknown): boolean {
  return isRecord(error) && error.name === "TimeoutError";
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
