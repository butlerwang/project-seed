# project-seed Frontend DX Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `useAuth()` hook, `withAuth` HOC, Sonner toast notifications, zod + react-hook-form validation in login/register forms, and a proper landing page template to project-seed's frontend.

**Architecture:** `useAuth()` encapsulates the `api.auth.me()` call + localStorage token + redirect logic so every page doesn't repeat the pattern. `withAuth` is a HOC that wraps pages needing authentication (replaces the manual `useEffect + router.push("/login")` in dashboard). Sonner's `<Toaster>` is added to the root layout once; individual pages call `toast.*()` instead of managing `error` state manually. Login and register forms are refactored to use zod schemas + react-hook-form, giving proper client-side validation. The landing page is upgraded from a 3-line stub to a real hero + features + CTA template ready for copy changes.

**Tech Stack:** Next.js 15 App Router, React 19, Tailwind, `sonner`, `react-hook-form`, `zod`, `@hookform/resolvers`

---

## File Map

| Action | Path | Responsibility |
|--------|------|---------------|
| Create | `frontend/lib/useAuth.ts` | Hook: loads user, redirects to /login if unauthenticated |
| Create | `frontend/lib/withAuth.tsx` | HOC: wraps pages behind auth gate |
| Modify | `frontend/app/layout.tsx` | Add `<Toaster>` from sonner |
| Modify | `frontend/app/(auth)/login/page.tsx` | Refactor to zod + react-hook-form + toast |
| Modify | `frontend/app/(auth)/register/page.tsx` | Refactor to zod + react-hook-form + toast |
| Modify | `frontend/app/dashboard/page.tsx` | Replace manual useEffect auth with `useAuth()` |
| Modify | `frontend/app/page.tsx` | Replace stub with hero + features + CTA landing page |
| Modify | `frontend/package.json` | Add sonner, react-hook-form, zod, @hookform/resolvers |

---

## Task 1: Install dependencies

**Files:**
- Modify: `frontend/package.json`

- [ ] **Step 1: Install packages**

```bash
cd frontend
npm install sonner react-hook-form zod @hookform/resolvers
```

- [ ] **Step 2: Verify install**

```bash
npx tsc --noEmit
```

Expected: no new errors (same as before).

- [ ] **Step 3: Commit**

```bash
git add package.json package-lock.json
git commit -m "chore(frontend): add sonner, react-hook-form, zod, @hookform/resolvers"
```

---

## Task 2: Add Sonner Toaster to root layout

**Files:**
- Modify: `frontend/app/layout.tsx`

- [ ] **Step 1: Update `frontend/app/layout.tsx`**

The current layout:
```typescript
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "project-seed",
  description: "Your SaaS description here",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;600;700&family=DM+Sans:ital,opsz,wght@0,9..40,300;0,9..40,400;0,9..40,500;1,9..40,300&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="font-sans antialiased bg-white text-gray-900">{children}</body>
    </html>
  );
}
```

Replace with:
```typescript
import type { Metadata } from "next";
import { Toaster } from "sonner";
import "./globals.css";

export const metadata: Metadata = {
  title: "project-seed",
  description: "Your SaaS description here",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@400;500;600;700&family=DM+Sans:ital,opsz,wght@0,9..40,300;0,9..40,400;0,9..40,500;1,9..40,300&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="font-sans antialiased bg-white text-gray-900">
        {children}
        <Toaster position="top-right" richColors />
      </body>
    </html>
  );
}
```

- [ ] **Step 2: Type check**

```bash
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add app/layout.tsx
git commit -m "feat(frontend): add Sonner Toaster to root layout"
```

---

## Task 3: useAuth hook

**Files:**
- Create: `frontend/lib/useAuth.ts`

- [ ] **Step 1: Implement `frontend/lib/useAuth.ts`**

```typescript
"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, type User } from "./api";

interface AuthState {
  user: User | null;
  loading: boolean;
}

/**
 * useAuth — loads the current user from /auth/me.
 * If the request fails (401 or network error), redirects to /login.
 * Returns { user, loading }.
 *
 * Usage:
 *   const { user, loading } = useAuth();
 *   if (loading) return <Spinner />;
 *   // user is guaranteed non-null here
 */
export function useAuth(): AuthState {
  const router = useRouter();
  const [state, setState] = useState<AuthState>({ user: null, loading: true });

  useEffect(() => {
    api.auth
      .me()
      .then((user) => setState({ user, loading: false }))
      .catch(() => {
        router.push("/login");
      });
  }, [router]);

  return state;
}
```

- [ ] **Step 2: Type check**

```bash
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add lib/useAuth.ts
git commit -m "feat(frontend): add useAuth() hook"
```

---

## Task 4: withAuth higher-order component

**Files:**
- Create: `frontend/lib/withAuth.tsx`

- [ ] **Step 1: Implement `frontend/lib/withAuth.tsx`**

```typescript
"use client";

import { useAuth } from "./useAuth";

/**
 * withAuth wraps a page component and renders null while the auth check loads.
 * If the user is not authenticated, useAuth() redirects to /login automatically.
 *
 * Usage:
 *   export default withAuth(DashboardPage);
 */
export function withAuth<P extends object>(
  Component: React.ComponentType<P & { user: NonNullable<ReturnType<typeof useAuth>["user"]> }>
): React.FC<P> {
  return function AuthGuard(props: P) {
    const { user, loading } = useAuth();
    if (loading || !user) return null;
    return <Component {...props} user={user} />;
  };
}
```

- [ ] **Step 2: Type check**

```bash
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add lib/withAuth.tsx
git commit -m "feat(frontend): add withAuth HOC"
```

---

## Task 5: Refactor Dashboard to use useAuth

**Files:**
- Modify: `frontend/app/dashboard/page.tsx`

The current dashboard has a manual `useEffect` that calls `api.auth.me()` and redirects. Replace it with `useAuth()`.

- [ ] **Step 1: Replace `frontend/app/dashboard/page.tsx`**

```typescript
"use client";

import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useAuth } from "@/lib/useAuth";
import { logout } from "@/lib/auth";

export default function DashboardPage() {
  const router = useRouter();
  const { user, loading } = useAuth();

  async function handleLogout() {
    await logout();
    toast.success("Signed out");
    router.push("/login");
  }

  if (loading || !user) return <div className="p-8 text-gray-400">Loading...</div>;

  return (
    <main className="p-8 max-w-3xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <h1 className="font-display text-3xl font-bold">Dashboard</h1>
        <button
          onClick={handleLogout}
          className="text-sm text-gray-500 hover:text-gray-900"
        >
          Sign out
        </button>
      </div>
      <div className="bg-white border border-gray-100 rounded-2xl p-6 shadow-sm">
        <p className="text-sm text-gray-500 mb-1">Signed in as</p>
        <p className="font-medium">{user.email}</p>
        <p className="text-sm text-gray-400 mt-1">Plan: {user.plan} · Role: {user.role}</p>
      </div>
    </main>
  );
}
```

- [ ] **Step 2: Type check**

```bash
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add app/dashboard/page.tsx
git commit -m "refactor(frontend): use useAuth() in dashboard, add logout toast"
```

---

## Task 6: Refactor login form with zod + react-hook-form + toast

**Files:**
- Modify: `frontend/app/(auth)/login/page.tsx`

- [ ] **Step 1: Replace `frontend/app/(auth)/login/page.tsx`**

```typescript
"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { login } from "@/lib/auth";

const schema = z.object({
  email: z.string().email("Enter a valid email"),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

type FormValues = z.infer<typeof schema>;

export default function LoginPage() {
  const router = useRouter();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  async function onSubmit(values: FormValues) {
    try {
      await login(values.email, values.password);
      router.push("/dashboard");
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Login failed");
    }
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
      <h1 className="font-display text-2xl font-bold mb-6">Sign in</h1>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1.5">Email</label>
          <input
            type="email"
            {...register("email")}
            className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          {errors.email && (
            <p className="text-xs text-red-500 mt-1">{errors.email.message}</p>
          )}
        </div>
        <div>
          <label className="block text-sm font-medium mb-1.5">Password</label>
          <input
            type="password"
            {...register("password")}
            className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          {errors.password && (
            <p className="text-xs text-red-500 mt-1">{errors.password.message}</p>
          )}
        </div>
        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full bg-gray-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-gray-700 transition-colors disabled:opacity-50"
        >
          {isSubmitting ? "Signing in..." : "Sign in"}
        </button>
      </form>
      <p className="text-sm text-gray-500 mt-4 text-center">
        No account?{" "}
        <Link href="/register" className="text-gray-900 font-medium hover:underline">
          Register
        </Link>
      </p>
      <p className="text-sm text-gray-500 mt-2 text-center">
        <Link href="/forgot-password" className="text-gray-500 hover:text-gray-900 hover:underline">
          Forgot password?
        </Link>
      </p>
    </div>
  );
}
```

- [ ] **Step 2: Type check**

```bash
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add "app/(auth)/login/page.tsx"
git commit -m "refactor(frontend): login form uses zod + react-hook-form + sonner toast"
```

---

## Task 7: Refactor register form with zod + react-hook-form + toast

**Files:**
- Modify: `frontend/app/(auth)/register/page.tsx`

First, read the current register page to see its structure, then replace.

- [ ] **Step 1: Read current register page**

```bash
cat app/(auth)/register/page.tsx
```

- [ ] **Step 2: Replace `frontend/app/(auth)/register/page.tsx`**

```typescript
"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { register as registerUser } from "@/lib/auth";

const schema = z
  .object({
    email: z.string().email("Enter a valid email"),
    password: z.string().min(8, "Password must be at least 8 characters"),
    confirmPassword: z.string(),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type FormValues = z.infer<typeof schema>;

export default function RegisterPage() {
  const router = useRouter();
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({ resolver: zodResolver(schema) });

  async function onSubmit(values: FormValues) {
    try {
      await registerUser(values.email, values.password);
      toast.success("Account created! Welcome aboard.");
      router.push("/dashboard");
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Registration failed");
    }
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
      <h1 className="font-display text-2xl font-bold mb-6">Create account</h1>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1.5">Email</label>
          <input
            type="email"
            {...register("email")}
            className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          {errors.email && (
            <p className="text-xs text-red-500 mt-1">{errors.email.message}</p>
          )}
        </div>
        <div>
          <label className="block text-sm font-medium mb-1.5">Password</label>
          <input
            type="password"
            {...register("password")}
            className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          {errors.password && (
            <p className="text-xs text-red-500 mt-1">{errors.password.message}</p>
          )}
        </div>
        <div>
          <label className="block text-sm font-medium mb-1.5">Confirm password</label>
          <input
            type="password"
            {...register("confirmPassword")}
            className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          {errors.confirmPassword && (
            <p className="text-xs text-red-500 mt-1">{errors.confirmPassword.message}</p>
          )}
        </div>
        <button
          type="submit"
          disabled={isSubmitting}
          className="w-full bg-gray-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-gray-700 transition-colors disabled:opacity-50"
        >
          {isSubmitting ? "Creating account..." : "Get started"}
        </button>
      </form>
      <p className="text-sm text-gray-500 mt-4 text-center">
        Already have an account?{" "}
        <Link href="/login" className="text-gray-900 font-medium hover:underline">
          Sign in
        </Link>
      </p>
    </div>
  );
}
```

- [ ] **Step 3: Type check**

```bash
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add "app/(auth)/register/page.tsx"
git commit -m "refactor(frontend): register form uses zod + react-hook-form + sonner toast"
```

---

## Task 8: Landing page — hero + features + CTA

**Files:**
- Modify: `frontend/app/page.tsx`

Replace the 5-line stub with a production-ready landing page template. Everything is placeholder copy — the project's `make rename` step will update the product name; copy changes are left for the team.

- [ ] **Step 1: Replace `frontend/app/page.tsx`**

```typescript
import Link from "next/link";

const features = [
  {
    title: "Feature one",
    description: "Describe the first core value prop in one or two sentences. What pain does this remove?",
    icon: "⚡",
  },
  {
    title: "Feature two",
    description: "The second differentiator. Keep it benefit-led, not capability-led.",
    icon: "🔒",
  },
  {
    title: "Feature three",
    description: "Third selling point. If you have more than three, you haven't prioritized yet.",
    icon: "📈",
  },
];

export default function LandingPage() {
  return (
    <main className="min-h-screen bg-white">
      {/* Nav */}
      <nav className="max-w-5xl mx-auto px-6 py-5 flex items-center justify-between">
        <span className="font-display text-lg font-bold tracking-tight">project-seed</span>
        <div className="flex items-center gap-4">
          <Link href="/login" className="text-sm text-gray-500 hover:text-gray-900">
            Sign in
          </Link>
          <Link
            href="/register"
            className="text-sm bg-gray-900 text-white px-4 py-2 rounded-lg font-medium hover:bg-gray-700 transition-colors"
          >
            Get started
          </Link>
        </div>
      </nav>

      {/* Hero */}
      <section className="max-w-4xl mx-auto px-6 pt-24 pb-20 text-center">
        <h1 className="font-display text-5xl sm:text-6xl font-bold tracking-tight leading-tight mb-6">
          Your product tagline
          <br />
          <span className="text-gray-400">goes right here.</span>
        </h1>
        <p className="text-xl text-gray-500 max-w-2xl mx-auto mb-10">
          One or two sentences explaining who this is for and what it does. Keep it
          concrete — avoid adjectives like "powerful" and "seamless."
        </p>
        <div className="flex gap-4 justify-center">
          <Link
            href="/register"
            className="px-8 py-3.5 bg-gray-900 text-white rounded-xl font-medium hover:bg-gray-700 transition-colors text-sm"
          >
            Start for free
          </Link>
          <Link
            href="#features"
            className="px-8 py-3.5 border border-gray-200 rounded-xl font-medium hover:bg-gray-50 transition-colors text-sm"
          >
            See how it works
          </Link>
        </div>
        <p className="text-xs text-gray-400 mt-4">No credit card required · Free forever on the starter plan</p>
      </section>

      {/* Social proof / logo bar placeholder */}
      <section className="border-y border-gray-100 py-10">
        <p className="text-center text-sm text-gray-400 mb-6">Trusted by teams at</p>
        <div className="flex items-center justify-center gap-10 opacity-40">
          {["Company A", "Company B", "Company C", "Company D"].map((name) => (
            <span key={name} className="text-sm font-semibold text-gray-600">
              {name}
            </span>
          ))}
        </div>
      </section>

      {/* Features */}
      <section id="features" className="max-w-5xl mx-auto px-6 py-24">
        <h2 className="font-display text-3xl font-bold text-center mb-4">Everything you need</h2>
        <p className="text-gray-500 text-center max-w-xl mx-auto mb-16">
          Replace this with a sentence that frames why these features matter together.
        </p>
        <div className="grid sm:grid-cols-3 gap-8">
          {features.map((f) => (
            <div key={f.title} className="rounded-2xl border border-gray-100 p-6 hover:shadow-sm transition-shadow">
              <div className="text-2xl mb-4">{f.icon}</div>
              <h3 className="font-display font-semibold mb-2">{f.title}</h3>
              <p className="text-sm text-gray-500 leading-relaxed">{f.description}</p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className="bg-gray-900 mx-6 mb-16 rounded-3xl px-10 py-16 text-center max-w-5xl lg:mx-auto">
        <h2 className="font-display text-3xl font-bold text-white mb-4">Ready to get started?</h2>
        <p className="text-gray-400 mb-8 max-w-md mx-auto text-sm">
          Join thousands of teams who ship faster with project-seed. Set up in under 5 minutes.
        </p>
        <Link
          href="/register"
          className="inline-block px-8 py-3.5 bg-white text-gray-900 rounded-xl font-medium hover:bg-gray-100 transition-colors text-sm"
        >
          Create free account
        </Link>
      </section>

      {/* Footer */}
      <footer className="max-w-5xl mx-auto px-6 pb-10 flex items-center justify-between text-xs text-gray-400">
        <span>© {new Date().getFullYear()} project-seed</span>
        <div className="flex gap-6">
          <Link href="/privacy" className="hover:text-gray-600">Privacy</Link>
          <Link href="/terms" className="hover:text-gray-600">Terms</Link>
        </div>
      </footer>
    </main>
  );
}
```

- [ ] **Step 2: Type check**

```bash
npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add app/page.tsx
git commit -m "feat(frontend): replace landing page stub with hero + features + CTA template"
```

---

## Task 9: Final type check and build

- [ ] **Step 1: Full type check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 2: Next.js build check**

```bash
npm run build
```

Expected: build completes with no errors. (Warnings about `<img>` vs `<Image>` in the landing page are acceptable — replace with `next/image` if you want Lighthouse points.)

- [ ] **Step 3: Visual check in browser**

```bash
npm run dev
```

Open:
- `http://localhost:3000` — landing page with hero + features + CTA
- `http://localhost:3000/login` — login form with inline validation errors
- `http://localhost:3000/register` — register form with password confirmation

Test:
1. Submit login with an invalid email — field error appears inline, no page reload
2. Submit login with short password — field error appears
3. Submit valid credentials — redirects to `/dashboard`
4. Sign out from dashboard — toast "Signed out" appears, redirect to `/login`

- [ ] **Step 4: Commit final**

```bash
git add .
git commit -m "chore(frontend): verify DX improvements build and type-check clean"
```
