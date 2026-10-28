import "server-only";

// One JSON line per entry on stdout, with the same keys as the Go API's slog JSON handler,
// so Alloy ships both services' logs to Loki and one `| json` query reads them.

export type LogAttrs = Record<string, string | number | boolean>;

type Level = "INFO" | "ERROR";

function write(level: Level, msg: string, attrs: LogAttrs): void {
  console.log(
    JSON.stringify({ time: new Date().toISOString(), level, msg, ...attrs }),
  );
}

export const logger = {
  info: (msg: string, attrs: LogAttrs = {}) => write("INFO", msg, attrs),
  error: (msg: string, attrs: LogAttrs = {}) => write("ERROR", msg, attrs),
};
