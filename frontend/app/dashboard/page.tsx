"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, type User } from "@/lib/api";
import { logout } from "@/lib/auth";

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    api.auth.me().then(setUser).catch(() => router.push("/login"));
  }, [router]);

  async function handleLogout() {
    await logout();
    router.push("/login");
  }

  if (!user) return <div className="p-8 text-gray-400">Loading...</div>;

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
