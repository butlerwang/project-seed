"use client";

import Link from "next/link";
import { useState } from "react";

import { api } from "@/lib/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    try {
      await api.auth.forgotPassword(email);
      setSent(true);
    } catch {
      setError("Something went wrong. Try again.");
    }
  }

  if (sent) {
    return (
      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8 text-center">
        <h1 className="font-display text-2xl font-bold mb-2">Check your email</h1>
        <p className="text-gray-500">If that address exists, we sent a reset link.</p>
        <Link href="/login" className="mt-4 inline-block text-sm text-gray-900 font-medium hover:underline">
          Back to login
        </Link>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
      <h1 className="font-display text-2xl font-bold mb-6">Forgot password</h1>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-1.5">Email</label>
          <input
            type="email"
            required
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            className="w-full border border-gray-200 rounded-lg px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
        </div>
        {error && <p className="text-sm text-red-600">{error}</p>}
        <button
          type="submit"
          className="w-full bg-gray-900 text-white rounded-lg py-2.5 text-sm font-medium hover:bg-gray-700 transition-colors"
        >
          Send reset link
        </button>
      </form>
      <p className="text-sm text-gray-500 mt-4 text-center">
        <Link href="/login" className="text-gray-900 font-medium hover:underline">
          Back to login
        </Link>
      </p>
    </div>
  );
}
