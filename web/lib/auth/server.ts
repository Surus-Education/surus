import { cookies } from "next/headers";
import type { User } from "@/lib/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1";

export async function getServerSession(): Promise<User | null> {
  const cookieStore = await cookies();
  const token = cookieStore.get("access_token");

  if (!token) return null;

  try {
    const res = await fetch(`${API_URL}/auth/me`, {
      headers: {
        "Content-Type": "application/json",
        Cookie: `access_token=${token.value}`,
      },
      cache: "no-store",
    });

    if (!res.ok) return null;
    const data = await res.json();
    return data.user;
  } catch {
    return null;
  }
}
