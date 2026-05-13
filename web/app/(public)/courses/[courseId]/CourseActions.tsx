"use client";

import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { saveCourse, unsaveCourse, forkCourse } from "@/lib/api/courses";
import { useAuth } from "@/hooks/useAuth";
import { useUIStore } from "@/lib/stores/uiStore";
import { ReportButton } from "@/components/shared/ReportButton";
import { Bookmark, GitFork } from "lucide-react";
import { toast } from "sonner";
import { useState } from "react";

export function CourseActions({ courseId }: { courseId: string }) {
  const { user } = useAuth();
  const { openAuthPrompt } = useUIStore();
  const router = useRouter();
  const [saved, setSaved] = useState(false);

  const saveMutation = useMutation({
    mutationFn: () => (saved ? unsaveCourse(courseId) : saveCourse(courseId)),
    onSuccess: () => {
      setSaved(!saved);
      toast.success(saved ? "Removed from library" : "Saved to library");
    },
  });

  const forkMutation = useMutation({
    mutationFn: () => forkCourse(courseId),
    onSuccess: (data) => {
      router.push(`/courses/${data.course.id}/edit`);
    },
    onError: () => {
      toast.error("Failed to fork course");
    },
  });

  const handleSave = () => {
    if (!user) {
      openAuthPrompt("save this course");
      return;
    }
    saveMutation.mutate();
  };

  const handleFork = () => {
    if (!user) {
      openAuthPrompt("fork this course");
      return;
    }
    forkMutation.mutate();
  };

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" onClick={handleSave}>
        <Bookmark className={`h-4 w-4 mr-1 ${saved ? "fill-current" : ""}`} />
        {saved ? "Saved" : "Save"}
      </Button>
      <Button variant="outline" size="sm" onClick={handleFork} disabled={forkMutation.isPending}>
        <GitFork className="h-4 w-4 mr-1" />
        {forkMutation.isPending ? "Forking..." : "Fork"}
      </Button>
      <ReportButton targetType="course" courseId={courseId} />
    </div>
  );
}
