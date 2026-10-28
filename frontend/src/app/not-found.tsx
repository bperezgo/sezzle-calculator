import Link from "next/link";

export default function NotFound() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 px-4 py-8 text-center">
      <h1 className="text-2xl font-semibold">Page not found</h1>
      <p className="text-zinc-600 dark:text-zinc-400">
        There&apos;s nothing at this address.
      </p>
      <Link
        href="/"
        className="flex min-h-11 items-center rounded-lg px-4 text-lg font-semibold text-blue-700 underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 dark:text-blue-400"
      >
        Back to the calculator
      </Link>
    </main>
  );
}
