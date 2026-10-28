import { describe, expect, it } from "vitest";

import { isDecimal, parseCalculationForm } from "./validation";

function form(fields: Record<string, string>): FormData {
  const data = new FormData();
  for (const [name, value] of Object.entries(fields)) data.set(name, value);
  return data;
}

describe("isDecimal", () => {
  it.each(["0", "-2", "+3", "2.5", ".5", "5.", "1e3", "-1.5E-7"])(
    "accepts %s",
    (value) => {
      expect(isDecimal(value)).toBe(true);
    },
  );

  it.each(["", "abc", "1,5", "1e", "--1", "Infinity", "NaN", "0x10", "1 2"])(
    "rejects %j",
    (value) => {
      expect(isDecimal(value)).toBe(false);
    },
  );
});

describe("parseCalculationForm", () => {
  it("builds a binary request, trimming operands", () => {
    expect(
      parseCalculationForm(form({ operation: "divide", a: " 10 ", b: "4" })),
    ).toEqual({
      ok: true,
      request: { operation: "divide", operands: ["10", "4"] },
    });
  });

  it("ignores the second operand for a unary operation", () => {
    expect(
      parseCalculationForm(form({ operation: "sqrt", a: "2", b: "junk" })),
    ).toEqual({ ok: true, request: { operation: "sqrt", operands: ["2"] } });
  });

  it("reports an unknown operation", () => {
    const parsed = parseCalculationForm(
      form({ operation: "modulo", a: "1", b: "2" }),
    );

    expect(parsed).toEqual({
      ok: false,
      fieldErrors: { operation: "unsupported_operation" },
    });
  });

  it("reports missing and invalid operands per field", () => {
    const parsed = parseCalculationForm(form({ operation: "add", b: "abc" }));

    expect(parsed).toEqual({
      ok: false,
      fieldErrors: { a: "required", b: "invalid_number" },
    });
  });

  it("treats a missing operation as required", () => {
    expect(parseCalculationForm(form({ a: "1", b: "2" }))).toEqual({
      ok: false,
      fieldErrors: { operation: "required" },
    });
  });
});
