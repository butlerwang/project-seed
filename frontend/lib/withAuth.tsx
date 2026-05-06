"use client";

import { useAuth } from "./useAuth";

export function withAuth<P extends object>(
  Component: React.ComponentType<P & { user: NonNullable<ReturnType<typeof useAuth>["user"]> }>
): React.FC<P> {
  return function AuthGuard(props: P) {
    const { user, loading } = useAuth();
    if (loading || !user) return null;
    return <Component {...props} user={user} />;
  };
}
