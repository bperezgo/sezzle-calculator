import { calculateAction } from "@/app/actions";
import { Calculator } from "@/components/calculator/calculator";

export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-6 px-4 py-8">
      <h1 className="text-3xl font-semibold">Calculator</h1>
      <div className="w-full max-w-sm">
        <Calculator action={calculateAction} />
      </div>
    </main>
  );
}
