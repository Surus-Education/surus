import type { PublicUser, Course } from "@/lib/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1";

export async function getUser(userId: string): Promise<{ user: PublicUser; courses: Course[] }> {
  const res = await fetch(`${API_URL}/users/${userId}`, {
    cache: "no-store",
    headers: { "Content-Type": "application/json" },
  });

  if (!res.ok) throw new Error(`API error: ${res.status}`);
  return res.json();
}
