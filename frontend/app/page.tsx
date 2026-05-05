import Link from "next/link";

export default function Home() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center gap-6 p-8">
      <h1 className="font-display text-5xl font-bold tracking-tight">project-seed</h1>
      <p className="text-gray-500 text-lg">Your product tagline here.</p>
      <div className="flex gap-4">
        <Link
          href="/login"
          className="px-6 py-3 bg-gray-900 text-white rounded-lg font-medium hover:bg-gray-700 transition-colors"
        >
          Sign in
        </Link>
        <Link
          href="/register"
          className="px-6 py-3 border border-gray-300 rounded-lg font-medium hover:bg-gray-50 transition-colors"
        >
          Get started
        </Link>
      </div>
    </main>
  );
}
