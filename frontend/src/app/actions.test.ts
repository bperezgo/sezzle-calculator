import { afterEach, describe, expect, it, vi } from "vitest";

import { calculate } from "@/lib/api/client";
import type { CalculationResponse } from "@/lib/api/types";
import { initialCalculatorState } from "@/lib/calculator-state";

import { calculateAction } from "./actions";

vi.mock("@/lib/api/client", () => ({ calculate: vi.fn() }));

const calculateMock = vi.mocked(calculate);

function form(fields: Record<string, string>): FormData {
  const data = new FormData();
  for (const [name, value] of Object.entries(fields)) data.set(name, value);
  return data;
}

afterEach(() => {
  calculateMock.mockReset();
});

describe("calculateAction", () => {
  it.each([
    ["an empty operand", { operation: "add", a: "", b: "2" }],
    ["a non-numeric operand", { operation: "add", a: "abc", b: "2" }],
    ["an unknown operation", { operation: "modulo", a: "1", b: "2" }],
    ["a missing operation", { a: "1", b: "2" }],
  ])("never calls the client for %s", async (_, fields) => {
    const state = await calculateAction(initialCalculatorState, form(fields));

    expect(calculateMock).not.toHaveBeenCalled();
    expect(state.status).toBe("error");
    expect(state.status === "error" && state.fieldErrors).not.toEqual({});
  });

  it("returns the field errors and keeps the inputs", async () => {
    const state = await calculateAction(
      initialCalculatorState,
      form({ operation: "divide", a: "abc", b: "" }),
    );

    expect(state).toEqual({
      status: "error",
      values: { operation: "divide", a: "abc", b: "" },
      fieldErrors: { a: "invalid_number", b: "required" },
    });
  });

  it("returns success with the result for valid input", async () => {
    const calculation: CalculationResponse = {
      operation: "divide",
      operands: ["10", "4"],
      result: "2.5",
    };
    calculateMock.mockResolvedValue({ ok: true, result: calculation });

    const state = await calculateAction(
      initialCalculatorState,
      form({ operation: "divide", a: "10", b: "4" }),
    );

    expect(calculateMock).toHaveBeenCalledWith({
      operation: "divide",
      operands: ["10", "4"],
    });
    expect(state).toEqual({
      status: "success",
      values: { operation: "divide", a: "10", b: "4" },
      calculation,
    });
  });

  it("returns the API error and keeps the inputs", async () => {
    const error = {
      code: "division_by_zero",
      message: "division by zero is undefined",
    } as const;
    calculateMock.mockResolvedValue({ ok: false, error });

    const state = await calculateAction(
      initialCalculatorState,
      form({ operation: "divide", a: "10", b: "0" }),
    );

    expect(state).toEqual({
      status: "error",
      values: { operation: "divide", a: "10", b: "0" },
      fieldErrors: {},
      error,
    });
  });
});
