import Link from "next/link";

const features = [
  {
    title: "Operational backbone included",
    description:
      "Start with authentication, billing hooks, storage, and LLM plumbing already wired so the team can focus on product logic instead of boilerplate.",
    accent: "from-amber-300 via-orange-300 to-rose-300",
  },
  {
    title: "Built to be renamed fast",
    description:
      "The template is intentionally generic where it should be and opinionated where velocity matters, so new projects get a clean launch surface instead of a dead scaffold.",
    accent: "from-sky-300 via-cyan-300 to-emerald-300",
  },
  {
    title: "Production-minded defaults",
    description:
      "Structured backend plumbing, guarded auth flows, and deploy-friendly frontend conventions give the next team a sharper baseline on day one.",
    accent: "from-fuchsia-300 via-pink-300 to-red-300",
  },
];

export default function LandingPage() {
  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top,#fff4db_0%,#fff8ef_28%,#fffdf9_58%,#ffffff_100%)] text-gray-900">
      <div className="absolute inset-x-0 top-0 -z-10 h-[34rem] bg-[linear-gradient(135deg,rgba(248,203,113,0.22),rgba(245,120,120,0.08)_42%,rgba(255,255,255,0)_72%)]" />

      <nav className="mx-auto flex max-w-6xl items-center justify-between px-6 py-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-gray-900 text-sm font-bold text-white shadow-[0_14px_40px_rgba(17,24,39,0.18)]">
            ps
          </div>
          <div>
            <p className="font-display text-lg font-bold tracking-tight">project-seed</p>
            <p className="text-xs uppercase tracking-[0.28em] text-gray-400">Launch template</p>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Link href="/login" className="text-sm text-gray-500 transition-colors hover:text-gray-900">
            Sign in
          </Link>
          <Link
            href="/register"
            className="rounded-full bg-gray-900 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-gray-700"
          >
            Start free
          </Link>
        </div>
      </nav>

      <section className="mx-auto grid max-w-6xl gap-12 px-6 pb-24 pt-14 lg:grid-cols-[1.15fr_0.85fr] lg:items-center">
        <div>
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-amber-200 bg-white/80 px-4 py-2 text-xs font-medium uppercase tracking-[0.24em] text-amber-700 shadow-sm backdrop-blur">
            SaaS boilerplate with real opinions
          </div>
          <h1 className="font-display text-5xl font-bold leading-[0.95] tracking-tight text-gray-950 sm:text-6xl lg:text-7xl">
            Ship the first
            <span className="block text-amber-600">serious version</span>
            without rebuilding the platform.
          </h1>
          <p className="mt-7 max-w-2xl text-lg leading-8 text-gray-600 sm:text-xl">
            project-seed gives a new product team a sharper starting line: auth, storage, LLM
            routing, deploy hooks, and a frontend shell that already feels like software instead of a demo.
          </p>
          <div className="mt-10 flex flex-col gap-4 sm:flex-row">
            <Link
              href="/register"
              className="rounded-2xl bg-gray-900 px-7 py-4 text-center text-sm font-medium text-white shadow-[0_18px_45px_rgba(17,24,39,0.2)] transition-transform hover:-translate-y-0.5 hover:bg-gray-700"
            >
              Create free account
            </Link>
            <Link
              href="/login"
              className="rounded-2xl border border-gray-200 bg-white/90 px-7 py-4 text-center text-sm font-medium text-gray-900 transition-colors hover:border-gray-300 hover:bg-white"
            >
              Explore the app surface
            </Link>
          </div>
          <div className="mt-10 grid gap-4 text-sm text-gray-500 sm:grid-cols-3">
            <div className="rounded-2xl border border-white/70 bg-white/70 p-4 backdrop-blur">
              <p className="font-display text-2xl font-bold text-gray-950">Auth</p>
              <p className="mt-1">Email verify, reset flows, OAuth-ready paths.</p>
            </div>
            <div className="rounded-2xl border border-white/70 bg-white/70 p-4 backdrop-blur">
              <p className="font-display text-2xl font-bold text-gray-950">LLM</p>
              <p className="mt-1">Provider router plus token streaming support.</p>
            </div>
            <div className="rounded-2xl border border-white/70 bg-white/70 p-4 backdrop-blur">
              <p className="font-display text-2xl font-bold text-gray-950">Ops</p>
              <p className="mt-1">Uploads, webhooks, rate limits, and deploy rails.</p>
            </div>
          </div>
        </div>

        <div className="relative">
          <div className="absolute -left-8 top-12 h-28 w-28 rounded-full bg-amber-200/70 blur-3xl" />
          <div className="absolute -right-10 bottom-10 h-36 w-36 rounded-full bg-rose-200/60 blur-3xl" />
          <div className="relative overflow-hidden rounded-[2rem] border border-white/80 bg-white/85 p-6 shadow-[0_28px_70px_rgba(17,24,39,0.12)] backdrop-blur">
            <div className="mb-5 flex items-center justify-between">
              <div>
                <p className="text-xs uppercase tracking-[0.24em] text-gray-400">Control room</p>
                <p className="font-display text-2xl font-semibold text-gray-950">Starter workspace</p>
              </div>
              <div className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700">
                Ready to clone
              </div>
            </div>
            <div className="space-y-4">
              <div className="rounded-2xl bg-gray-950 p-5 text-white">
                <p className="text-xs uppercase tracking-[0.24em] text-white/50">Flow</p>
                <p className="mt-2 font-display text-2xl font-semibold">Register → onboard → stream results</p>
                <p className="mt-2 text-sm leading-6 text-white/70">
                  The core routes are already there, so the first real demo can be about your product instead of your plumbing.
                </p>
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="rounded-2xl border border-gray-100 p-5">
                  <p className="text-xs uppercase tracking-[0.24em] text-gray-400">Backend</p>
                  <p className="mt-2 text-sm leading-6 text-gray-600">
                    Structured logs, storage abstraction, Stripe webhook stub, and extensible auth.
                  </p>
                </div>
                <div className="rounded-2xl border border-gray-100 p-5">
                  <p className="text-xs uppercase tracking-[0.24em] text-gray-400">Frontend</p>
                  <p className="mt-2 text-sm leading-6 text-gray-600">
                    App Router shell, protected dashboard pattern, validated forms, and streaming-ready hooks.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="features" className="mx-auto max-w-6xl px-6 pb-24">
        <div className="mb-12 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-xs uppercase tracking-[0.26em] text-gray-400">Why this starter exists</p>
            <h2 className="font-display mt-3 text-3xl font-bold tracking-tight text-gray-950 sm:text-4xl">
              The boring platform work is already moving.
            </h2>
          </div>
          <p className="max-w-xl text-sm leading-7 text-gray-500">
            Rename the product, swap the copy, and start shaping the business logic. The template is meant to disappear under your product, not fight it.
          </p>
        </div>
        <div className="grid gap-6 lg:grid-cols-3">
          {features.map((feature) => (
            <article
              key={feature.title}
              className="group relative overflow-hidden rounded-[1.75rem] border border-gray-100 bg-white p-7 shadow-[0_16px_40px_rgba(17,24,39,0.06)] transition-transform hover:-translate-y-1"
            >
              <div
                className={`mb-6 h-12 w-12 rounded-2xl bg-gradient-to-br ${feature.accent} shadow-inner`}
              />
              <h3 className="font-display text-2xl font-semibold text-gray-950">{feature.title}</h3>
              <p className="mt-3 text-sm leading-7 text-gray-600">{feature.description}</p>
              <div className="mt-8 h-px w-full bg-gradient-to-r from-transparent via-gray-200 to-transparent" />
              <p className="mt-5 text-xs uppercase tracking-[0.22em] text-gray-400">
                Ship faster, rename later
              </p>
            </article>
          ))}
        </div>
      </section>

      <section className="mx-auto max-w-6xl px-6 pb-20">
        <div className="overflow-hidden rounded-[2rem] bg-gray-950 px-8 py-12 text-white shadow-[0_26px_70px_rgba(17,24,39,0.22)] sm:px-12">
          <div className="flex flex-col gap-8 lg:flex-row lg:items-center lg:justify-between">
            <div className="max-w-2xl">
              <p className="text-xs uppercase tracking-[0.28em] text-white/45">Next step</p>
              <h2 className="font-display mt-4 text-3xl font-bold tracking-tight sm:text-4xl">
                Start with the real thing, not a placeholder sprint.
              </h2>
              <p className="mt-4 text-sm leading-7 text-white/70">
                Use this baseline to stand up the first working version, then replace the template language as product truth gets sharper.
              </p>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Link
                href="/register"
                className="rounded-full bg-white px-6 py-3 text-center text-sm font-medium text-gray-900 transition-colors hover:bg-gray-100"
              >
                Create account
              </Link>
              <Link
                href="/login"
                className="rounded-full border border-white/20 px-6 py-3 text-center text-sm font-medium text-white/90 transition-colors hover:border-white/40 hover:text-white"
              >
                Sign in
              </Link>
            </div>
          </div>
        </div>
      </section>
    </main>
  );
}
