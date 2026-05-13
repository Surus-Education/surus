import { apiFetch } from "./client";
import type { PublicUser, User, Course } from "@/lib/types";

export async function getUser(
  userId: string
): Promise<{ user: PublicUser }> {
  return apiFetch(`/users/${userId}`);
}

export async function updateMe(input: {
  display_name?: string;
  bio?: string;
  avatar_url?: string;
}): Promise<{ user: User }> {
  return apiFetch("/users/me", {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export async function getLibrary(): Promise<{
  saved: Course[];
  created: Course[];
}> {
  return apiFetch("/users/me/library");
}

export async function deleteMe(): Promise<void> {
  return apiFetch("/users/me", { method: "DELETE" });
}

export async function restoreMe(): Promise<void> {
  return apiFetch("/users/me/restore", { method: "POST" });
}

export async function exportData(): Promise<void> {
  return apiFetch("/users/me/export", { method: "POST" });
}
