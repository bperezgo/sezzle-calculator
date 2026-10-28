import type { ApiErrorCode } from "@/lib/api/types";
import type { CalculatorField, FieldErrorCode } from "@/lib/validation";

// Every user-facing error message lives here, keyed by the stable codes. The API's own
// `message` is for developers and never shown.

const SOMETHING_WENT_WRONG = "Something went wrong. Try again.";

const API_ERROR_MESSAGES: Record<ApiErrorCode, string> = {
  invalid_request: SOMETHING_WENT_WRONG,
  unsupported_operation: "Choose an operation from the list.",
  invalid_operand_count: SOMETHING_WENT_WRONG,
  invalid_operand: "Enter valid numbers.",
  operand_too_large: "Numbers can have at most 50 significant digits.",
  payload_too_large: "The numbers you entered are too long.",
  division_by_zero: "You can't divide by zero.",
  undefined_power: "That power is undefined for these numbers.",
  negative_square_root: "You can't take the square root of a negative number.",
  result_too_large: "The result is too large to calculate.",
  not_found: SOMETHING_WENT_WRONG,
  method_not_allowed: SOMETHING_WENT_WRONG,
  internal_error: SOMETHING_WENT_WRONG,
  backend_unavailable: "The calculator service is unavailable. Try again.",
};

// The operand an API error is about, so the UI can flag that input.
const API_ERROR_FIELDS: Partial<Record<ApiErrorCode, CalculatorField>> = {
  division_by_zero: "b",
  negative_square_root: "a",
};

export function apiErrorMessage(code: ApiErrorCode): string {
  return API_ERROR_MESSAGES[code];
}

export function apiErrorField(code: ApiErrorCode): CalculatorField | undefined {
  return API_ERROR_FIELDS[code];
}

export function fieldErrorMessage(
  field: CalculatorField,
  code: FieldErrorCode,
): string {
  switch (code) {
    case "required":
      return field === "operation" ? "Choose an operation" : "Enter a number";
    case "invalid_number":
      return "Enter a valid number";
    case "unsupported_operation":
      return "Choose an operation from the list";
  }
}
