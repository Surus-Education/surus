import { apiFetch } from "./client";
import type { Course, CourseInput, PaginatedResponse, ReportInput, Report } from "@/lib/types";

export async function getCourses(params?: {
  q?: string;
  tags?: string;
  cursor?: string;
  limit?: number;
}): Promise<PaginatedResponse<Course>> {
  const search = new URLSearchParams();
  if (params?.q) search.set("q", params.q);
  if (params?.tags) search.set("tags", params.tags);
  if (params?.cursor) search.set("cursor", params.cursor);
  if (params?.limit) search.set("limit", params.limit.toString());
  const qs = search.toString();
  return apiFetch(`/courses${qs ? `?${qs}` : ""}`);
}

export async function getCourse(
  courseId: string
): Promise<{ course: Course }> {
  return apiFetch(`/courses/${courseId}`);
}

export async function createCourse(
  input: CourseInput
): Promise<{ course: Course }> {
  return apiFetch("/courses", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateCourse(
  courseId: string,
  input: Partial<CourseInput>
): Promise<{ course: Course }> {
  return apiFetch(`/courses/${courseId}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export async function deleteCourse(courseId: string): Promise<void> {
  return apiFetch(`/courses/${courseId}`, { method: "DELETE" });
}

export async function forkCourse(
  courseId: string
): Promise<{ course: Course }> {
  return apiFetch(`/courses/${courseId}/fork`, { method: "POST" });
}

export async function saveCourse(courseId: string): Promise<void> {
  return apiFetch(`/courses/${courseId}/save`, { method: "POST" });
}

export async function unsaveCourse(courseId: string): Promise<void> {
  return apiFetch(`/courses/${courseId}/save`, { method: "DELETE" });
}

export async function reportCourse(
  courseId: string,
  input: ReportInput
): Promise<{ report: Report }> {
  return apiFetch(`/courses/${courseId}/report`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}
