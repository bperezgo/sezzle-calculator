import { describe, expect, it } from "vitest";

import { API_ERROR_CODES } from "@/lib/api/types";

import { apiErrorField, apiErrorMessage, fieldErrorMessage } from "./errors";

describe("fieldErrorMessage", () => {
  it.each([
    ["a", "required", "Enter a number"],
    ["b", "required", "Enter a number"],
    ["a", "invalid_number", "Enter a valid number"],
    ["operation", "required", "Choose an operation"],
    ["operation", "unsupported_operation", "Choose an operation from the list"],
  ] as const)("%s %s → %j", (field, code, message) => {
    expect(fieldErrorMessage(field, code)).toBe(message);
  });
});

describe("apiErrorMessage", () => {
  it.each([
    ["division_by_zero", "You can't divide by zero."],
    [
      "negative_square_root",
      "You can't take the square root of a negative number.",
    ],
    [
      "backend_unavailable",
      "The calculator service is unavailable. Try again.",
    ],
    ["internal_error", "Something went wrong. Try again."],
  ] as const)("%s → %j", (code, message) => {
    expect(apiErrorMessage(code)).toBe(message);
  });

  it.each([...API_ERROR_CODES, "backend_unavailable" as const])(
    "has a user-facing message for %s",
    (code) => {
      const message = apiErrorMessage(code);
      expect(message).toMatch(/^[A-Z].*\.$/);
      expect(message).not.toContain("_");
    },
  );
});

describe("apiErrorField", () => {
  it.each([
    ["division_by_zero", "b"],
    ["negative_square_root", "a"],
    ["backend_unavailable", undefined],
    ["result_too_large", undefined],
  ] as const)("%s → %s", (code, field) => {
    expect(apiErrorField(code)).toBe(field);
  });
});
