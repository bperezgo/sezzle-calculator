import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { logger } from "./logger";

const consoleLog = vi.spyOn(console, "log");

beforeEach(() => {
  consoleLog.mockImplementation(() => {});
});

afterEach(() => {
  consoleLog.mockReset();
});

function lastLine(): unknown {
  expect(consoleLog).toHaveBeenCalledTimes(1);
  const [line] = consoleLog.mock.calls[0];
  expect(typeof line).toBe("string");
  expect(String(line)).not.toContain("\n");
  return JSON.parse(String(line));
}

describe("logger", () => {
  it.each([
    ["info", "INFO"],
    ["error", "ERROR"],
  ] as const)(
    "%s writes one JSON line with the slog keys and the attrs",
    (method, level) => {
      logger[method]("backend call", { request_id: "r-1", status: 200 });

      const entry = lastLine();
      expect(entry).toEqual({
        time: expect.any(String),
        level,
        msg: "backend call",
        request_id: "r-1",
        status: 200,
      });
      const { time } = entry as { time: string };
      expect(new Date(time).toISOString()).toBe(time);
    },
  );

  it("works without attrs", () => {
    logger.info("started");

    expect(lastLine()).toMatchObject({ level: "INFO", msg: "started" });
  });
});
