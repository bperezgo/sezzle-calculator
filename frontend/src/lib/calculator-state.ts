import type { ApiError, CalculationResponse } from "@/lib/api/types";
import type { CalculatorField, FieldErrors } from "@/lib/validation";

// Serialisable state for useActionState. Lives outside actions.ts because a "use server"
// module may only export async functions. `values` echoes the inputs so they survive errors.
export type CalculatorValues = Record<CalculatorField, string>;

export type CalculatorState =
  | { status: "idle"; values: CalculatorValues }
  | {
      status: "success";
      values: CalculatorValues;
      calculation: CalculationResponse;
    }
  | {
      status: "error";
      values: CalculatorValues;
      fieldErrors: FieldErrors;
      error?: ApiError;
    };

export const initialCalculatorState: CalculatorState = {
  status: "idle",
  values: { operation: "add", a: "", b: "" },
};
