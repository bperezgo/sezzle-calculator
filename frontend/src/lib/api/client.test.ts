import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { calculate } from "./client";
import type { ApiErrorCode, CalculationRequest } from "./types";

const BACKEND = "http://backend.test";
const request: CalculationRequest = {
  operation: "divide",
  operands: ["1", "4"],
};

const fetchMock = vi.fn<typeof fetch>();
const consoleLog = vi.spyOn(console, "log");

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json; charset=utf-8" },
  });
}

const unavailable = {
  ok: false,
  error: { code: "backend_unavailable", message: expect.any(String) },
};

beforeEach(() => {
  vi.stubEnv("BACKEND_URL", BACKEND);
  vi.stubGlobal("fetch", fetchMock);
  consoleLog.mockImplementation(() => {});
});

afterEach(() => {
  fetchMock.mockReset();
  consoleLog.mockReset();
  vi.unstubAllEnvs();
  vi.unstubAllGlobals();
});

describe("calculate", () => {
  it("posts the request and returns the calculation on success", async () => {
    const body = { operation: "divide", operands: ["1", "4"], result: "0.25" };
    fetchMock.mockResolvedValue(jsonResponse(200, body));

    await expect(calculate(request)).resolves.toEqual({
      ok: true,
      result: body,
    });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe(`${BACKEND}/api/v1/calculations`);
    expect(init?.method).toBe("POST");
    expect(init?.cache).toBe("no-store");
    expect(new Headers(init?.headers).get("Content-Type")).toBe(
      "application/json",
    );
    expect(JSON.parse(String(init?.body))).toEqual(request);
    expect(init?.signal).toBeInstanceOf(AbortSignal);
  });

  it("keeps the result as an exact decimal string", async () => {
    const result = "0.3333333333333333";
    fetchMock.mockResolvedValue(
      jsonResponse(200, { operation: "divide", operands: ["1", "3"], result }),
    );

    const outcome = await calculate({
      operation: "divide",
      operands: ["1", "3"],
    });

    expect(outcome.ok && outcome.result.result).toBe(result);
  });

  it.each<[ApiErrorCode, number]>([
    ["invalid_request", 400],
    ["unsupported_operation", 400],
    ["invalid_operand_count", 400],
    ["invalid_operand", 400],
    ["operand_too_large", 400],
    ["payload_too_large", 413],
    ["division_by_zero", 422],
    ["undefined_power", 422],
    ["negative_square_root", 422],
    ["result_too_large", 422],
    ["not_found", 404],
    ["method_not_allowed", 405],
  ])("returns the API error %s (%i)", async (code, status) => {
    const message = `message for ${code}`;
    fetchMock.mockResolvedValue(
      jsonResponse(status, { error: { code, message } }),
    );

    await expect(calculate(request)).resolves.toEqual({
      ok: false,
      error: { code, message },
    });
  });

  it.each([
    [
      "500 internal_error",
      () =>
        jsonResponse(500, {
          error: { code: "internal_error", message: "boom" },
        }),
    ],
    [
      "502 with an HTML body",
      () => new Response("<html>Bad Gateway</html>", { status: 502 }),
    ],
    ["503 with no body", () => new Response(null, { status: 503 })],
  ])("maps a %s to backend_unavailable", async (_, response) => {
    fetchMock.mockResolvedValue(response());

    await expect(calculate(request)).resolves.toEqual(unavailable);
  });

  it.each([
    ["malformed JSON", () => new Response("{not json", { status: 200 })],
    [
      "a success body missing the result",
      () => jsonResponse(200, { operation: "divide", operands: ["1", "4"] }),
    ],
    [
      "a numeric result",
      () =>
        jsonResponse(200, {
          operation: "divide",
          operands: ["1", "4"],
          result: 0.25,
        }),
    ],
    [
      "an unknown operation",
      () =>
        jsonResponse(200, {
          operation: "modulo",
          operands: ["1", "4"],
          result: "1",
        }),
    ],
    [
      "an error without the envelope",
      () => jsonResponse(400, { message: "bad" }),
    ],
    [
      "an unknown error code",
      () => jsonResponse(400, { error: { code: "teapot", message: "?" } }),
    ],
    ["malformed JSON on an error", () => new Response("oops", { status: 400 })],
  ])("maps %s to backend_unavailable", async (_, response) => {
    fetchMock.mockResolvedValue(response());

    await expect(calculate(request)).resolves.toEqual(unavailable);
  });

  it("maps a network error to backend_unavailable", async () => {
    fetchMock.mockRejectedValue(new TypeError("fetch failed"));

    await expect(calculate(request)).resolves.toEqual(unavailable);
  });

  it("aborts after the timeout and maps it to backend_unavailable", async () => {
    fetchMock.mockImplementation(
      (_, init) =>
        new Promise((_, reject) => {
          init?.signal?.addEventListener("abort", () =>
            reject(init.signal?.reason),
          );
        }),
    );

    await expect(calculate(request, { timeoutMs: 10 })).resolves.toEqual(
      unavailable,
    );
    expect(fetchMock.mock.calls[0][1]?.signal?.aborted).toBe(true);
  });
});

function logLines(): Record<string, unknown>[] {
  return consoleLog.mock.calls.map(([line]) => JSON.parse(String(line)));
}

function sentRequestId(call = 0): string | null {
  return new Headers(fetchMock.mock.calls[call][1]?.headers).get(
    "X-Request-ID",
  );
}

describe("calculate logging", () => {
  const success = () =>
    jsonResponse(200, {
      operation: "divide",
      operands: ["1", "4"],
      result: "0.25",
    });

  it("sends a unique X-Request-ID the backend accepts", async () => {
    fetchMock.mockImplementation(async () => success());

    await calculate(request);
    await calculate(request);

    const [first, second] = [sentRequestId(0), sentRequestId(1)];
    expect(first).toMatch(/^[A-Za-z0-9._-]{1,128}$/);
    expect(second).toMatch(/^[A-Za-z0-9._-]{1,128}$/);
    expect(first).not.toBe(second);
  });

  it("logs a success at INFO with the request ID it sent", async () => {
    fetchMock.mockResolvedValue(success());

    await calculate(request);

    expect(logLines()).toEqual([
      {
        time: expect.any(String),
        level: "INFO",
        msg: "backend call",
        request_id: sentRequestId(),
        operation: "divide",
        status: 200,
        outcome: "ok",
        duration_ms: expect.any(Number),
      },
    ]);
  });

  it("logs an API error at INFO with its code", async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(422, {
        error: { code: "division_by_zero", message: "division by zero" },
      }),
    );

    await calculate(request);

    expect(logLines()).toEqual([
      expect.objectContaining({
        level: "INFO",
        status: 422,
        outcome: "api_error",
        error_code: "division_by_zero",
      }),
    ]);
  });

  it.each<[string, () => void, Record<string, unknown>]>([
    [
      "network",
      () => fetchMock.mockRejectedValue(new TypeError("fetch failed")),
      { reason: "network" },
    ],
    [
      "timeout",
      () =>
        fetchMock.mockRejectedValue(
          new DOMException("The operation timed out.", "TimeoutError"),
        ),
      { reason: "timeout" },
    ],
    [
      "server_error",
      () => fetchMock.mockResolvedValue(new Response(null, { status: 503 })),
      { reason: "server_error", status: 503 },
    ],
    [
      "invalid_response",
      () =>
        fetchMock.mockResolvedValue(new Response("{not json", { status: 200 })),
      { reason: "invalid_response", status: 200 },
    ],
    [
      "invalid_response (unknown error code)",
      () =>
        fetchMock.mockResolvedValue(
          jsonResponse(400, { error: { code: "teapot", message: "?" } }),
        ),
      { reason: "invalid_response", status: 400 },
    ],
  ])(
    "logs backend_unavailable at ERROR with reason %s",
    async (_, arrange, attrs) => {
      arrange();

      await expect(calculate(request)).resolves.toEqual(unavailable);

      expect(logLines()).toEqual([
        expect.objectContaining({
          level: "ERROR",
          msg: "backend call",
          request_id: sentRequestId(),
          operation: "divide",
          outcome: "backend_unavailable",
          ...attrs,
        }),
      ]);
    },
  );

  it("never logs the operands or the result", async () => {
    const secret: CalculationRequest = {
      operation: "add",
      operands: ["98765.4321", "12345.6789"],
    };
    fetchMock.mockResolvedValue(
      jsonResponse(200, { ...secret, result: "111111.111" }),
    );

    await calculate(secret);

    const line = String(consoleLog.mock.calls[0][0]);
    expect(line).not.toContain("98765.4321");
    expect(line).not.toContain("12345.6789");
    expect(line).not.toContain("111111.111");
  });
});
