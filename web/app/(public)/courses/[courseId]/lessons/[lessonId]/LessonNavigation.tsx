"use client";

import Link from "next/link";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { markComplete, unmarkComplete } from "@/lib/api/lessons";
import { useAuth } from "@/hooks/useAuth";
import { useUIStore } from "@/lib/stores/uiStore";
import { ReportButton } from "@/components/shared/ReportButton";
import { ChevronLeft, ChevronRight, CheckCircle } from "lucide-react";
import { toast } from "sonner";
import { useState } from "react";
import type { Lesson } from "@/lib/types";

export function LessonNavigation({
  courseId,
  lessonId,
  prevLesson,
  nextLesson,
}: {
  courseId: string;
  lessonId: string;
  prevLesson: Lesson | null;
  nextLesson: Lesson | null;
}) {
  const { user } = useAuth();
  const { openAuthPrompt } = useUIStore();
  const [completed, setCompleted] = useState(false);

  const completeMutation = useMutation({
    mutationFn: () =>
      completed
        ? unmarkComplete(courseId, lessonId)
        : markComplete(courseId, lessonId),
    onSuccess: () => {
      setCompleted(!completed);
      toast.success(completed ? "Unmarked as complete" : "Marked as complete");
    },
  });

  const handleComplete = () => {
    if (!user) {
      openAuthPrompt("mark this lesson as complete");
      return;
    }
    completeMutation.mutate();
  };

  return (
    <div className="border-t pt-6 space-y-4">
      <div className="flex items-center justify-between">
        <Button
          variant="outline"
          size="sm"
          onClick={handleComplete}
          disabled={completeMutation.isPending}
        >
          <CheckCircle className={`h-4 w-4 mr-1 ${completed ? "text-green-600" : ""}`} />
          {completed ? "Completed" : "Mark as complete"}
        </Button>
        <ReportButton targetType="lesson" courseId={courseId} lessonId={lessonId} />
      </div>
      <div className="flex items-center justify-between">
        {prevLesson ? (
          <Button variant="ghost" size="sm" asChild>
            <Link href={`/courses/${courseId}/lessons/${prevLesson.id}`}>
              <ChevronLeft className="h-4 w-4 mr-1" />
              {prevLesson.title}
            </Link>
          </Button>
        ) : (
          <div />
        )}
        {nextLesson ? (
          <Button variant="ghost" size="sm" asChild>
            <Link href={`/courses/${courseId}/lessons/${nextLesson.id}`}>
              {nextLesson.title}
              <ChevronRight className="h-4 w-4 ml-1" />
            </Link>
          </Button>
        ) : (
          <div />
        )}
      </div>
    </div>
  );
}
