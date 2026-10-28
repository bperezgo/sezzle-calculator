import "server-only";

export function parseBackendUrl(raw: string | undefined): string {
  if (!raw) {
    throw new Error(
      "BACKEND_URL is not set. Point it at the Go API, e.g. BACKEND_URL=http://localhost:8080 (see frontend/.env.example).",
    );
  }
  let url: URL;
  try {
    url = new URL(raw);
  } catch {
    throw new Error(`BACKEND_URL is not a valid URL: "${raw}".`);
  }
  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new Error(`BACKEND_URL must use http or https: "${raw}".`);
  }
  return url.href.replace(/\/+$/, "");
}

let backendUrlCache: string | undefined;

export function backendUrl(): string {
  backendUrlCache ??= parseBackendUrl(process.env.BACKEND_URL);
  return backendUrlCache;
}
