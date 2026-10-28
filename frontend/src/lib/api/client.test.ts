import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { calculate } from "./client";
import type { ApiErrorCode, CalculationRequest } from "./types";

const BACKEND = "http://backend.test";
const request: CalculationRequest = {
  operation: "divide",
  operands: ["1", "4"],
};

const fetchMock = vi.fn<typeof fetch>();

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
});

afterEach(() => {
  fetchMock.mockReset();
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
