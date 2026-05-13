"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getMe, logout as apiLogout } from "@/lib/api/auth";
import type { User } from "@/lib/types";

export function useAuth(): { user: User | null; isLoading: boolean; logout: () => Promise<void> } {
  const queryClient = useQueryClient();

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

  const logout = async () => {
    await apiLogout();
    queryClient.setQueryData(["auth", "me"], null);
    queryClient.invalidateQueries({ queryKey: ["auth"] });
  };

  return { user: data ?? null, isLoading, logout };
}
