"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api, type User } from "@/lib/api";

export default function AdminPage() {
  const router = useRouter();
  const [users, setUsers] = useState<User[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    api.admin.listUsers().then((r) => setUsers(r.users)).catch((e: unknown) => {
      const err = e as { status?: number; message?: string };
      if (err.status === 401) router.push("/login");
      else if (err.status === 403) setError("Admin access required");
      else setError("Failed to load users");
    });
  }, [router]);

  if (error) return <div className="p-8 text-red-600">{error}</div>;

  return (
    <main className="p-8 max-w-4xl mx-auto">
      <h1 className="font-display text-3xl font-bold mb-8">Admin — Users</h1>
      <div className="bg-white border border-gray-100 rounded-2xl shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-100">
            <tr>
              <th className="text-left px-4 py-3 text-gray-500 font-medium">Email</th>
              <th className="text-left px-4 py-3 text-gray-500 font-medium">Role</th>
              <th className="text-left px-4 py-3 text-gray-500 font-medium">Plan</th>
              <th className="text-left px-4 py-3 text-gray-500 font-medium">Joined</th>
            </tr>
          </thead>
          <tbody>
            {users.map((u) => (
              <tr key={u.id} className="border-b border-gray-50 hover:bg-gray-50">
                <td className="px-4 py-3">{u.email}</td>
                <td className="px-4 py-3 text-gray-500">{u.role}</td>
                <td className="px-4 py-3 text-gray-500">{u.plan}</td>
                <td className="px-4 py-3 text-gray-400">{new Date(u.created_at).toLocaleDateString()}</td>
              </tr>
            ))}
            {users.length === 0 && (
              <tr><td colSpan={4} className="px-4 py-8 text-center text-gray-400">No users</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </main>
  );
}
