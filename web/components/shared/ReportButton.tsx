"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Flag } from "lucide-react";
import { reportCourse } from "@/lib/api/courses";
import { reportLesson } from "@/lib/api/lessons";
import { useAuth } from "@/hooks/useAuth";
import { useUIStore } from "@/lib/stores/uiStore";
import { toast } from "sonner";
import { ApiError } from "@/lib/api/client";

const CATEGORIES = [
  { value: "incorrect", label: "Incorrect content" },
  { value: "harmful", label: "Harmful content" },
  { value: "copyright", label: "Copyright violation" },
  { value: "other", label: "Other" },
] as const;

export function ReportButton({
  targetType,
  courseId,
  lessonId,
}: {
  targetType: "course" | "lesson";
  courseId: string;
  lessonId?: string;
}) {
  const { user } = useAuth();
  const { openAuthPrompt } = useUIStore();
  const [open, setOpen] = useState(false);
  const [category, setCategory] = useState<string>("incorrect");
  const [body, setBody] = useState("");

  const mutation = useMutation({
    mutationFn: async () => {
      if (targetType === "course") {
        return reportCourse(courseId, { category: category as any, body });
      }
      return reportLesson(courseId, lessonId!, { category: category as any, body });
    },
    onSuccess: () => {
      toast.success("Report submitted. Thank you.");
      setOpen(false);
      setBody("");
    },
    onError: (err) => {
      if (err instanceof ApiError && err.code === "conflict") {
        toast.info("You've already reported this.");
      } else {
        toast.error("Failed to submit report.");
      }
    },
  });

  const handleClick = () => {
    if (!user) {
      openAuthPrompt("report this content");
      return;
    }
    setOpen(true);
  };

  return (
    <>
      <Button variant="ghost" size="sm" onClick={handleClick}>
        <Flag className="h-4 w-4 mr-1" />
        Report
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Report content</DialogTitle>
            <DialogDescription>Help us understand the issue.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>Category</Label>
              <div className="space-y-1">
                {CATEGORIES.map((cat) => (
                  <label key={cat.value} className="flex items-center gap-2 text-sm">
                    <input
                      type="radio"
                      name="category"
                      value={cat.value}
                      checked={category === cat.value}
                      onChange={(e) => setCategory(e.target.value)}
                    />
                    {cat.label}
                  </label>
                ))}
              </div>
            </div>
            <div className="space-y-2">
              <Label>Details (optional)</Label>
              <Textarea
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder="Tell us more..."
                maxLength={2000}
              />
            </div>
            <Button
              onClick={() => mutation.mutate()}
              disabled={mutation.isPending}
              className="w-full"
            >
              {mutation.isPending ? "Submitting..." : "Submit report"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
