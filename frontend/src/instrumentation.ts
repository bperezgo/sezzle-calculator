// Runs once at server start: a missing or invalid BACKEND_URL stops the server
// instead of failing on the first calculation.
export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    const { backendUrl } = await import("@/lib/env");
    backendUrl();
  }
}
