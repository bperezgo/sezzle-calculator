"use client";

import {
  startTransition,
  useActionState,
  useState,
  type FormEvent,
} from "react";

import { OPERATION_ARITY, OPERATIONS, type Operation } from "@/lib/api/types";
import {
  initialCalculatorState,
  type CalculatorState,
} from "@/lib/calculator-state";
import { formatDecimal } from "@/lib/format";

import { describeCalculation, OPERATION_LABELS } from "./operations";

export type CalculatorAction = (
  state: CalculatorState,
  formData: FormData,
) => Promise<CalculatorState>;

type CalculatorProps = {
  action: CalculatorAction;
  initialState?: CalculatorState;
};

const field =
  "min-h-11 w-full rounded-lg border border-zinc-300 bg-white px-3 text-lg text-zinc-900 " +
  "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 " +
  "dark:border-zinc-700 dark:bg-zinc-900 dark:text-zinc-100 dark:focus-visible:outline-blue-400";

const label = "mb-1 block text-sm font-medium text-zinc-700 dark:text-zinc-300";

export function Calculator({
  action,
  initialState = initialCalculatorState,
}: CalculatorProps) {
  const [state, formAction, pending] = useActionState(action, initialState);
  const [operation, setOperation] = useState(state.values.operation);
  const [a, setA] = useState(state.values.a);
  const [b, setB] = useState(state.values.b);

  // A form action resets the form when it settles, and React 19 doesn't restore a
  // controlled <select>. Submitting manually skips the reset; without JS `action` still posts.
  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    startTransition(() => formAction(formData));
  }

  const unary =
    operation in OPERATION_ARITY &&
    OPERATION_ARITY[operation as Operation] === 1;

  return (
    <form
      action={formAction}
      onSubmit={handleSubmit}
      className="flex w-full flex-col gap-4 rounded-2xl border border-zinc-200 bg-zinc-50 p-4 shadow-sm sm:p-6 dark:border-zinc-800 dark:bg-zinc-950"
    >
      <div>
        <label htmlFor="a" className={label}>
          {unary ? "Number" : "First number"}
        </label>
        <input
          id="a"
          name="a"
          inputMode="decimal"
          autoComplete="off"
          value={a}
          onChange={(event) => setA(event.target.value)}
          className={field}
        />
      </div>

      <div>
        <label htmlFor="operation" className={label}>
          Operation
        </label>
        <select
          id="operation"
          name="operation"
          value={operation}
          onChange={(event) => setOperation(event.target.value)}
          className={field}
        >
          {OPERATIONS.map((op) => (
            <option key={op} value={op}>
              {`${OPERATION_LABELS[op].symbol} ${OPERATION_LABELS[op].name}`}
            </option>
          ))}
        </select>
      </div>

      {!unary && (
        <div>
          <label htmlFor="b" className={label}>
            Second number
          </label>
          <input
            id="b"
            name="b"
            inputMode="decimal"
            autoComplete="off"
            value={b}
            onChange={(event) => setB(event.target.value)}
            className={field}
          />
        </div>
      )}

      <button
        type="submit"
        disabled={pending}
        className="min-h-11 rounded-lg bg-blue-600 px-4 text-lg font-semibold text-white hover:bg-blue-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 disabled:cursor-wait disabled:opacity-60 dark:bg-blue-500 dark:hover:bg-blue-400 dark:focus-visible:outline-blue-400"
      >
        {pending ? "Calculating…" : "Calculate"}
      </button>

      <output
        role="status"
        aria-live="polite"
        htmlFor="a operation b"
        className="block min-h-16 rounded-lg bg-white p-3 dark:bg-zinc-900"
      >
        <ResultPanel state={state} />
      </output>
    </form>
  );
}

function ResultPanel({ state }: { state: CalculatorState }) {
  switch (state.status) {
    case "idle":
      return null;
    case "success":
      return (
        <>
          <span className="block text-sm break-all text-zinc-600 dark:text-zinc-400">
            {`${describeCalculation(state.calculation)} =`}
          </span>{" "}
          <span
            title={state.calculation.result}
            className="block max-h-48 overflow-y-auto font-mono text-2xl font-semibold break-all"
          >
            {formatDecimal(state.calculation.result)}
          </span>
        </>
      );
    case "error":
      return (
        <span className="block text-red-700 dark:text-red-400">
          {state.error?.message ?? "Check the numbers you entered."}
        </span>
      );
  }
}
