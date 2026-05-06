"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import { api, type User } from "./api";

interface AuthState {
  user: User | null;
  loading: boolean;
}

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
