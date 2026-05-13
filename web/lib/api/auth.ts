import { apiFetch } from "./client";
import type { User } from "@/lib/types";

export async function getMe(): Promise<{ user: User }> {
  return apiFetch("/auth/me");
}

export async function logout(): Promise<void> {
  return apiFetch("/auth/logout", { method: "POST" });
}

export async function requestMagicLink(email: string): Promise<void> {
  return apiFetch("/auth/magic-link/request", {
    method: "POST",
    body: JSON.stringify({ email }),
  });
}

export async function verifyMagicLink(token: string): Promise<{ user: User }> {
  return apiFetch("/auth/magic-link/verify", {
    method: "POST",
    body: JSON.stringify({ token }),
  });
}
