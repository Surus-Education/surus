import type { ApiErrorBody } from "@/lib/types";

export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public status: number
  ) {
    super(message);
    this.name = "ApiError";
  }
}

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1";

let isRefreshing = false;

export async function apiFetch<T>(
  path: string,
  init?: RequestInit
): Promise<T> {
  const url = `${API_URL}${path}`;
  const res = await fetch(url, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (res.status === 401 && !isRefreshing) {
    isRefreshing = true;
    const refreshRes = await fetch(`${API_URL}/auth/refresh`, {
      method: "POST",
      credentials: "include",
    });
    isRefreshing = false;

    if (refreshRes.ok) {
      return apiFetch(path, init);
    }
  }

  if (!res.ok) {
    let body: ApiErrorBody;
    try {
      body = await res.json();
    } catch {
      throw new ApiError("internal_error", "An unexpected error occurred", res.status);
    }
    throw new ApiError(body.error.code, body.error.message, res.status);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return res.json();
}
