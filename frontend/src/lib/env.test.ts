import { describe, expect, it } from "vitest";

import { parseBackendUrl } from "./env";

describe("parseBackendUrl", () => {
  it.each([
    ["http://localhost:8080", "http://localhost:8080"],
    ["http://backend:8080/", "http://backend:8080"],
    ["https://api.example.com/base/", "https://api.example.com/base"],
  ])("accepts %s", (raw, expected) => {
    expect(parseBackendUrl(raw)).toBe(expected);
  });

  it.each([
    ["missing", undefined],
    ["empty", ""],
    ["not a URL", "localhost:8080"],
    ["unsupported scheme", "ftp://backend:8080"],
  ])("fails fast with a clear message when %s", (_, raw) => {
    expect(() => parseBackendUrl(raw)).toThrow(/BACKEND_URL/);
  });
});
