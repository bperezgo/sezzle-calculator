"use server";

import { calculate } from "@/lib/api/client";
import type { CalculatorState } from "@/lib/calculator-state";
import { formValue, parseCalculationForm } from "@/lib/validation";

export async function calculateAction(
  _prevState: CalculatorState,
  formData: FormData,
): Promise<CalculatorState> {
  const values = {
    operation: formValue(formData, "operation"),
    a: formValue(formData, "a"),
    b: formValue(formData, "b"),
  };

  const parsed = parseCalculationForm(formData);
  if (!parsed.ok) {
    return { status: "error", values, fieldErrors: parsed.fieldErrors };
  }

  const outcome = await calculate(parsed.request);
  if (!outcome.ok) {
    return { status: "error", values, fieldErrors: {}, error: outcome.error };
  }

  return { status: "success", values, calculation: outcome.result };
}
