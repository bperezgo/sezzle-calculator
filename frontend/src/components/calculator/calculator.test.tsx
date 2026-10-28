import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import {
  initialCalculatorState,
  type CalculatorState,
} from "@/lib/calculator-state";

import { Calculator, type CalculatorAction } from "./calculator";

function success(
  operation: "divide" | "sqrt",
  operands: string[],
  result: string,
): CalculatorState {
  const [a, b = ""] = operands;
  return {
    status: "success",
    values: { operation, a, b },
    calculation: { operation, operands, result },
  };
}

describe("Calculator", () => {
  it("renders every operation in the picker", () => {
    render(<Calculator action={vi.fn<CalculatorAction>()} />);

    const picker = screen.getByRole("combobox", { name: "Operation" });
    const options = within(picker)
      .getAllByRole("option")
      .map((option) => option.textContent);

    expect(options).toEqual([
      "+ Add",
      "− Subtract",
      "× Multiply",
      "÷ Divide",
      "^ Power",
      "√ Square root",
      "% Percentage",
    ]);
  });

  it("labels every input", () => {
    render(<Calculator action={vi.fn<CalculatorAction>()} />);

    expect(screen.getByLabelText("First number")).toBeInTheDocument();
    expect(screen.getByLabelText("Second number")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Calculate" }),
    ).toBeInTheDocument();
  });

  it("hides the second operand for a unary operation", async () => {
    const user = userEvent.setup();
    render(<Calculator action={vi.fn<CalculatorAction>()} />);

    await user.selectOptions(screen.getByLabelText("Operation"), "sqrt");

    expect(screen.getByLabelText("Number")).toBeInTheDocument();
    expect(screen.queryByLabelText("Second number")).not.toBeInTheDocument();

    await user.selectOptions(screen.getByLabelText("Operation"), "add");

    expect(screen.getByLabelText("Second number")).toBeInTheDocument();
  });

  it("entering 10 ÷ 4 and submitting shows 2.5", async () => {
    const user = userEvent.setup();
    const action = vi.fn<CalculatorAction>(async () =>
      success("divide", ["10", "4"], "2.5"),
    );
    render(<Calculator action={action} />);

    await user.type(screen.getByLabelText("First number"), "10");
    await user.selectOptions(screen.getByLabelText("Operation"), "÷ Divide");
    await user.type(screen.getByLabelText("Second number"), "4");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    const output = await screen.findByRole("status");
    expect(output).toHaveTextContent("10 ÷ 4 = 2.5");
    expect(screen.getByLabelText("First number")).toHaveValue("10");
    expect(screen.getByLabelText("Operation")).toHaveValue("divide");
    expect(screen.getByLabelText("Second number")).toHaveValue("4");
    const formData = action.mock.calls[0][1];
    expect(Object.fromEntries(formData)).toEqual({
      operation: "divide",
      a: "10",
      b: "4",
    });
  });

  it("submits with the Enter key", async () => {
    const user = userEvent.setup();
    const action = vi.fn<CalculatorAction>(async () =>
      success("divide", ["10", "4"], "2.5"),
    );
    render(<Calculator action={action} />);

    await user.selectOptions(screen.getByLabelText("Operation"), "divide");
    await user.type(screen.getByLabelText("First number"), "10");
    await user.type(screen.getByLabelText("Second number"), "4{Enter}");

    expect(await screen.findByRole("status")).toHaveTextContent("2.5");
  });

  it("shows the result from the action state, grouped, with the exact value in its title", () => {
    render(
      <Calculator
        action={vi.fn<CalculatorAction>()}
        initialState={success("divide", ["1234567", "1"], "1234567.5")}
      />,
    );

    const result = screen.getByText("1,234,567.5");
    expect(result).toHaveAttribute("title", "1234567.5");
    expect(screen.getByRole("status")).toHaveAttribute("aria-live", "polite");
  });

  it("keeps the submitted inputs from the action state", () => {
    render(
      <Calculator
        action={vi.fn<CalculatorAction>()}
        initialState={success("sqrt", ["16"], "4")}
      />,
    );

    expect(screen.getByLabelText("Operation")).toHaveValue("sqrt");
    expect(screen.getByLabelText("Number")).toHaveValue("16");
    expect(screen.getByRole("status")).toHaveTextContent("√16 = 4");
  });

  it("shows an error from the action state", () => {
    render(
      <Calculator
        action={vi.fn<CalculatorAction>()}
        initialState={{
          status: "error",
          values: { operation: "divide", a: "1", b: "0" },
          fieldErrors: {},
          error: {
            code: "division_by_zero",
            message: "division by zero is undefined",
          },
        }}
      />,
    );

    expect(screen.getByRole("status")).toHaveTextContent(
      "division by zero is undefined",
    );
  });

  it("disables submit while the calculation is pending", async () => {
    const user = userEvent.setup();
    let finish: (state: CalculatorState) => void = () => {};
    const action = vi.fn<CalculatorAction>(
      () =>
        new Promise((resolve) => {
          finish = resolve;
        }),
    );
    render(<Calculator action={action} />);

    await user.type(screen.getByLabelText("First number"), "1");
    await user.type(screen.getByLabelText("Second number"), "2");
    await user.click(screen.getByRole("button", { name: "Calculate" }));

    const button = await screen.findByRole("button", { name: "Calculating…" });
    expect(button).toBeDisabled();

    finish(initialCalculatorState);

    expect(
      await screen.findByRole("button", { name: "Calculate" }),
    ).toBeEnabled();
  });
});
