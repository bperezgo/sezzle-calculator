import {
  OPERATION_ARITY,
  OPERATIONS,
  type CalculationRequest,
  type Operation,
} from "@/lib/api/types";

export type CalculatorField = "operation" | "a" | "b";

export type FieldErrorCode =
  "required" | "invalid_number" | "unsupported_operation";

export type FieldErrors = Partial<Record<CalculatorField, FieldErrorCode>>;

export type ParsedCalculationForm =
  | { ok: true; request: CalculationRequest }
  | { ok: false; fieldErrors: FieldErrors };

// Same syntax the backend accepts (sign, optional fraction, exponent). Size limits stay
// with the backend, which is the final authority.
const DECIMAL = /^[+-]?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$/;

export function isDecimal(value: string): boolean {
  return DECIMAL.test(value);
}

export function formValue(formData: FormData, field: CalculatorField): string {
  const value = formData.get(field);
  return typeof value === "string" ? value.trim() : "";
}

function isOperation(value: string): value is Operation {
  return (OPERATIONS as readonly string[]).includes(value);
}

function operandError(value: string): FieldErrorCode | undefined {
  if (value === "") return "required";
  if (!isDecimal(value)) return "invalid_number";
  return undefined;
}

export function parseCalculationForm(
  formData: FormData,
): ParsedCalculationForm {
  const operation = formValue(formData, "operation");
  if (operation === "") {
    return { ok: false, fieldErrors: { operation: "required" } };
  }
  if (!isOperation(operation)) {
    return { ok: false, fieldErrors: { operation: "unsupported_operation" } };
  }

  const fields: CalculatorField[] =
    OPERATION_ARITY[operation] === 1 ? ["a"] : ["a", "b"];
  const operands = fields.map((field) => formValue(formData, field));

  const fieldErrors: FieldErrors = {};
  fields.forEach((field, i) => {
    const error = operandError(operands[i]);
    if (error) fieldErrors[field] = error;
  });
  if (Object.keys(fieldErrors).length > 0) return { ok: false, fieldErrors };

  return { ok: true, request: { operation, operands } };
}
