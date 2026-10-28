import { describe, expect, it } from "vitest";

import { formatDecimal } from "./format";

describe("formatDecimal", () => {
  it.each([
    ["2.5", "2.5"],
    ["0", "0"],
    ["999", "999"],
    ["1000", "1,000"],
    ["-1234567.89", "-1,234,567.89"],
    ["0.3333333333333333", "0.3333333333333333"],
    ["1234.5678", "1,234.5678"],
  ])("formats %s as %s", (value, expected) => {
    expect(formatDecimal(value)).toBe(expected);
  });

  it("keeps every digit of a 1001-digit integer", () => {
    const value = "1" + "0".repeat(1000);

    const formatted = formatDecimal(value);

    expect(formatted.replaceAll(",", "")).toBe(value);
    expect(formatted.startsWith("10,000,")).toBe(true);
  });

  it("never rounds a long fraction", () => {
    const value = "0." + "0".repeat(999) + "1";

    expect(formatDecimal(value)).toBe(value);
  });

  it("returns anything that is not a plain decimal unchanged", () => {
    expect(formatDecimal("1e3")).toBe("1e3");
  });
});
