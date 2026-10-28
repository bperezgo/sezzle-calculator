import type { CalculationResponse, Operation } from "@/lib/api/types";
import { formatDecimal } from "@/lib/format";

export const OPERATION_LABELS: Record<
  Operation,
  { symbol: string; name: string }
> = {
  add: { symbol: "+", name: "Add" },
  subtract: { symbol: "−", name: "Subtract" },
  multiply: { symbol: "×", name: "Multiply" },
  divide: { symbol: "÷", name: "Divide" },
  power: { symbol: "^", name: "Power" },
  sqrt: { symbol: "√", name: "Square root" },
  percentage: { symbol: "%", name: "Percentage" },
};

export function describeCalculation({
  operation,
  operands,
}: CalculationResponse): string {
  const [a, b] = operands.map(formatDecimal);
  switch (operation) {
    case "sqrt":
      return `√${a}`;
    case "percentage":
      return `${a}% of ${b}`;
    default:
      return `${a} ${OPERATION_LABELS[operation].symbol} ${b}`;
  }
}
