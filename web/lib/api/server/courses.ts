import type { Course, Lesson, PaginatedResponse } from "@/lib/types";
import { cookies } from "next/headers";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1";

async function serverFetch<T>(path: string, options?: { auth?: boolean; revalidate?: number }): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (options?.auth) {
    const cookieStore = await cookies();
    const token = cookieStore.get("access_token");
    if (token) {
      headers["Cookie"] = `access_token=${token.value}`;
    }
  }

  const res = await fetch(`${API_URL}${path}`, {
    headers,
    cache: options?.revalidate ? undefined : "no-store",
    next: options?.revalidate ? { revalidate: options.revalidate } : undefined,
  });

  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export async function getCourse(courseId: string): Promise<{ course: Course }> {
  return serverFetch(`/courses/${courseId}`, { auth: true });
}

export async function getCourses(params?: { q?: string; limit?: number }): Promise<PaginatedResponse<Course>> {
  const search = new URLSearchParams();
  if (params?.q) search.set("q", params.q);
  if (params?.limit) search.set("limit", params.limit.toString());
  const qs = search.toString();
  return serverFetch(`/courses${qs ? `?${qs}` : ""}`, { revalidate: 60 });
}

export async function getLessons(courseId: string): Promise<{ lessons: Lesson[] }> {
  return serverFetch(`/courses/${courseId}/lessons`, { auth: true });
}

export async function getLesson(courseId: string, lessonId: string): Promise<{ lesson: Lesson }> {
  return serverFetch(`/courses/${courseId}/lessons/${lessonId}`, { auth: true });
}
