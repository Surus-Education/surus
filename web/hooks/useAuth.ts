"use client";

import { useQuery } from "@tanstack/react-query";
import { getMe } from "@/lib/api/auth";
import type { User } from "@/lib/types";

export function useAuth(): { user: User | null; isLoading: boolean; logout: () => void } {
  const { data, isLoading } = useQuery({
    queryKey: ["auth", "me"],
    queryFn: async () => {
      try {
        const { user } = await getMe();
        return user;
      } catch {
        return null;
      }
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  });

  // Sign out is a top-level navigation to a Route Handler, not a fetch —
  // see app/auth/logout/route.ts. A fetch here could not set cookies visible
  // to the browser's own jar for this origin.
  //
  // It clears the web-origin jar and revokes refresh tokens + handoff codes
  // server-side, then redirects to "/". It does NOT clear the API-origin
  // cookie: that call is made from the Next server, so the API's deletion
  // Set-Cookie headers land on the server's response and never reach the
  // browser. The stateless access_token JWT therefore survives in the
  // API-origin jar until it expires (900s), during which client-side
  // apiFetch calls still authenticate. Tracked as PLAN.md 2.7d.
  const logout = () => {
    window.location.href = "/auth/logout";
  };

  return { user: data ?? null, isLoading, logout };
}
