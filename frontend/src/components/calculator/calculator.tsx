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
import {
  apiErrorField,
  apiErrorMessage,
  fieldErrorMessage,
} from "@/lib/errors";
import { formatDecimal } from "@/lib/format";
import {
  parseCalculationForm,
  type CalculatorField,
  type FieldErrors,
} from "@/lib/validation";

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

const invalidField =
  "aria-invalid:border-red-600 dark:aria-invalid:border-red-400";

const RESULT_MESSAGE_ID = "result-message";

function fieldErrorsOf(state: CalculatorState): FieldErrors {
  return state.status === "error" ? state.fieldErrors : {};
}

export function Calculator({
  action,
  initialState = initialCalculatorState,
}: CalculatorProps) {
  const [state, formAction, pending] = useActionState(action, initialState);
  const [operation, setOperation] = useState(state.values.operation);
  const [a, setA] = useState(state.values.a);
  const [b, setB] = useState(state.values.b);
  const [fieldErrors, setFieldErrors] = useState(fieldErrorsOf(state));
  const [shownState, setShownState] = useState(state);

  // Field errors come from the client check or, without JS, from the action's state.
  if (state !== shownState) {
    setShownState(state);
    setFieldErrors(fieldErrorsOf(state));
  }

  // A form action resets the form when it settles, and React 19 doesn't restore a
  // controlled <select>. Submitting manually skips the reset; without JS `action` still posts.
  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const formData = new FormData(form);
    const parsed = parseCalculationForm(formData);
    if (!parsed.ok) {
      setFieldErrors(parsed.fieldErrors);
      const first = (["a", "operation", "b"] as const).find(
        (name) => parsed.fieldErrors[name],
      );
      if (first) form.querySelector<HTMLElement>(`#${first}`)?.focus();
      return;
    }
    setFieldErrors({});
    startTransition(() => formAction(formData));
  }

  function edit(name: CalculatorField, setValue: (value: string) => void) {
    return (value: string) => {
      setValue(value);
      if (fieldErrors[name]) {
        setFieldErrors((errors) => ({ ...errors, [name]: undefined }));
      }
    };
  }

  const apiField =
    state.status === "error" && state.error
      ? apiErrorField(state.error.code)
      : undefined;

  function errorFor(name: CalculatorField) {
    const code = fieldErrors[name];
    if (code) {
      return {
        message: fieldErrorMessage(name, code),
        describedBy: `${name}-error`,
      };
    }
    if (apiField === name) return { describedBy: RESULT_MESSAGE_ID };
    return undefined;
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
      <OperandField
        name="a"
        label={unary ? "Number" : "First number"}
        value={a}
        onChange={edit("a", setA)}
        error={errorFor("a")}
      />

      <div>
        <label htmlFor="operation" className={label}>
          Operation
        </label>
        <select
          id="operation"
          name="operation"
          value={operation}
          onChange={(event) =>
            edit("operation", setOperation)(event.target.value)
          }
          aria-invalid={errorFor("operation") ? true : undefined}
          aria-describedby={errorFor("operation")?.describedBy}
          className={`${field} ${invalidField}`}
        >
          {OPERATIONS.map((op) => (
            <option key={op} value={op}>
              {`${OPERATION_LABELS[op].symbol} ${OPERATION_LABELS[op].name}`}
            </option>
          ))}
        </select>
        <FieldError
          id="operation-error"
          message={errorFor("operation")?.message}
        />
      </div>

      {!unary && (
        <OperandField
          name="b"
          label="Second number"
          value={b}
          onChange={edit("b", setB)}
          error={errorFor("b")}
        />
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

type FieldErrorProps = { message?: string; describedBy: string };

function OperandField({
  name,
  label: text,
  value,
  onChange,
  error,
}: {
  name: "a" | "b";
  label: string;
  value: string;
  onChange: (value: string) => void;
  error?: FieldErrorProps;
}) {
  return (
    <div>
      <label htmlFor={name} className={label}>
        {text}
      </label>
      <input
        id={name}
        name={name}
        inputMode="decimal"
        autoComplete="off"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        aria-invalid={error ? true : undefined}
        aria-describedby={error?.describedBy}
        className={`${field} ${invalidField}`}
      />
      <FieldError id={`${name}-error`} message={error?.message} />
    </div>
  );
}

function FieldError({ id, message }: { id: string; message?: string }) {
  if (!message) return null;
  return (
    <p id={id} className="mt-1 text-sm text-red-700 dark:text-red-400">
      {message}
    </p>
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
      if (!state.error) return null;
      return (
        <span
          id={RESULT_MESSAGE_ID}
          className="block text-red-700 dark:text-red-400"
        >
          {apiErrorMessage(state.error.code)}
        </span>
      );
  }
}
