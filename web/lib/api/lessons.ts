import { apiFetch } from "./client";
import type { Lesson, LessonInput, ReorderMove, QuizAttemptResult, ReportInput, Report } from "@/lib/types";

export async function getLessons(
  courseId: string
): Promise<{ lessons: Lesson[] }> {
  return apiFetch(`/courses/${courseId}/lessons`);
}

export async function getLesson(
  courseId: string,
  lessonId: string
): Promise<{ lesson: Lesson }> {
  return apiFetch(`/courses/${courseId}/lessons/${lessonId}`);
}

export async function createLesson(
  courseId: string,
  input: LessonInput
): Promise<{ lesson: Lesson }> {
  return apiFetch(`/courses/${courseId}/lessons`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateLesson(
  courseId: string,
  lessonId: string,
  input: Partial<LessonInput>
): Promise<{ lesson: Lesson }> {
  return apiFetch(`/courses/${courseId}/lessons/${lessonId}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export async function deleteLesson(
  courseId: string,
  lessonId: string
): Promise<void> {
  return apiFetch(`/courses/${courseId}/lessons/${lessonId}`, {
    method: "DELETE",
  });
}

export async function reorderLessons(
  courseId: string,
  moves: ReorderMove[]
): Promise<void> {
  return apiFetch(`/courses/${courseId}/lessons/reorder`, {
    method: "PATCH",
    body: JSON.stringify({ moves }),
  });
}

export async function markComplete(
  courseId: string,
  lessonId: string
): Promise<void> {
  return apiFetch(`/courses/${courseId}/lessons/${lessonId}/complete`, {
    method: "POST",
  });
}

export async function unmarkComplete(
  courseId: string,
  lessonId: string
): Promise<void> {
  return apiFetch(`/courses/${courseId}/lessons/${lessonId}/complete`, {
    method: "DELETE",
  });
}

export async function submitQuizAttempt(
  courseId: string,
  lessonId: string,
  answers: Record<string, string | string[]>
): Promise<QuizAttemptResult> {
  return apiFetch(`/courses/${courseId}/lessons/${lessonId}/attempts`, {
    method: "POST",
    body: JSON.stringify({ answers }),
  });
}

export async function reportLesson(
  courseId: string,
  lessonId: string,
  input: ReportInput
): Promise<{ report: Report }> {
  return apiFetch(`/courses/${courseId}/lessons/${lessonId}/report`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}
