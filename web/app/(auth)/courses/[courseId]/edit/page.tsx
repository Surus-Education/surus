"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { LessonOutlineEditor } from "@/components/editor/LessonOutlineEditor";
import { TiptapEditor } from "@/components/editor/TiptapEditor";
import { ForkBanner } from "@/components/course/ForkBanner";
import { getCourse, updateCourse } from "@/lib/api/courses";
import { getLessons, createLesson, updateLesson, deleteLesson, reorderLessons } from "@/lib/api/lessons";
import type { Lesson, TiptapDoc } from "@/lib/types";
import { toast } from "sonner";
import Link from "next/link";
import { Eye } from "lucide-react";

function parseYouTubeUrl(url: string): string | null {
  const patterns = [
    /(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})/,
  ];
  for (const p of patterns) {
    const m = url.match(p);
    if (m) return m[1];
  }
  return null;
}

export default function EditCoursePage() {
  const { courseId } = useParams<{ courseId: string }>();

  const { data: courseData } = useQuery({
    queryKey: ["courses", courseId],
    queryFn: () => getCourse(courseId),
  });

  const { data: lessonsData, refetch: refetchLessons } = useQuery({
    queryKey: ["courses", courseId, "lessons"],
    queryFn: () => getLessons(courseId),
  });

  const course = courseData?.course;
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [visibility, setVisibility] = useState<string>("private");
  const [saveStatus, setSaveStatus] = useState<"saved" | "saving" | "idle">("idle");
  const saveTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => {
    if (course) {
      setTitle(course.title);
      setDescription(course.description);
      setVisibility(course.visibility);
    }
  }, [course]);

  useEffect(() => {
    if (lessonsData?.lessons) {
      setLessons(lessonsData.lessons);
    }
  }, [lessonsData]);

  const selectedLesson = lessons.find((l) => l.id === selectedId);

  const debouncedSaveCourse = useCallback(
    (updates: any) => {
      if (saveTimeoutRef.current) clearTimeout(saveTimeoutRef.current);
      setSaveStatus("saving");
      saveTimeoutRef.current = setTimeout(async () => {
        try {
          await updateCourse(courseId, updates);
          setSaveStatus("saved");
        } catch {
          toast.error("Failed to save");
          setSaveStatus("idle");
        }
      }, 2000);
    },
    [courseId]
  );

  const handleAddLesson = async (type: "video" | "page" | "quiz") => {
    try {
      const input: any = {
        type,
        title: `New ${type} lesson`,
        position: lessons.length,
      };
      if (type === "page") input.page = { content: { type: "doc", content: [{ type: "paragraph" }] } };
      if (type === "quiz") input.quiz = { questions: [] };
      if (type === "video") input.video = { provider: "youtube", provider_id: "placeholder", source_url: "https://youtube.com" };

      const { lesson } = await createLesson(courseId, input);
      setLessons([...lessons, lesson]);
      setSelectedId(lesson.id);
    } catch {
      toast.error("Failed to add lesson");
    }
  };

  const handleDeleteLesson = async (id: string) => {
    try {
      await deleteLesson(courseId, id);
      setLessons(lessons.filter((l) => l.id !== id));
      if (selectedId === id) setSelectedId(null);
    } catch {
      toast.error("Failed to delete lesson");
    }
  };

  const handleReorder = (newLessons: Lesson[]) => {
    setLessons(newLessons);
    const moves = newLessons.map((l, i) => ({
      lesson_id: l.id,
      new_parent_id: l.parent_id,
      new_position: i,
    }));
    reorderLessons(courseId, moves).catch(() => toast.error("Failed to save order"));
  };

  if (!course) {
    return <div className="mx-auto max-w-7xl px-4 py-8">Loading...</div>;
  }

  return (
    <div className="flex h-[calc(100vh-3.5rem)]">
      <div className="w-72 border-r p-4 overflow-y-auto space-y-4">
        {course.forked_from_id && (
          <ForkBanner forkedFromId={course.forked_from_id} forkedAt={course.forked_at!} />
        )}
        <div>
          <Input
            value={title}
            onChange={(e) => {
              setTitle(e.target.value);
              debouncedSaveCourse({ title: e.target.value });
            }}
            placeholder="Course title"
            className="font-semibold"
          />
        </div>
        <div>
          <Select value={visibility} onValueChange={(v) => {
            setVisibility(v);
            debouncedSaveCourse({ visibility: v });
          }}>
            <SelectTrigger className="text-xs"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="private">Private</SelectItem>
              <SelectItem value="unlisted">Unlisted</SelectItem>
              <SelectItem value="public">Public</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" asChild>
            <Link href={`/courses/${courseId}`}>
              <Eye className="h-3.5 w-3.5 mr-1" /> Preview
            </Link>
          </Button>
          <span className="text-xs text-muted-foreground">
            {saveStatus === "saving" ? "Saving..." : saveStatus === "saved" ? "Saved" : ""}
          </span>
        </div>
        <LessonOutlineEditor
          lessons={lessons}
          selectedId={selectedId}
          onSelect={setSelectedId}
          onReorder={handleReorder}
          onAdd={handleAddLesson}
          onDelete={handleDeleteLesson}
        />
      </div>

      <div className="flex-1 p-6 overflow-y-auto">
        {selectedLesson ? (
          <LessonEditor
            courseId={courseId}
            lesson={selectedLesson}
            onUpdate={(updated) => {
              setLessons(lessons.map((l) => (l.id === updated.id ? updated : l)));
            }}
          />
        ) : (
          <div className="flex items-center justify-center h-full text-muted-foreground">
            Select a lesson to edit or add a new one.
          </div>
        )}
      </div>
    </div>
  );
}

function LessonEditor({
  courseId,
  lesson,
  onUpdate,
}: {
  courseId: string;
  lesson: Lesson;
  onUpdate: (lesson: Lesson) => void;
}) {
  const [title, setTitle] = useState(lesson.title);
  const [videoUrl, setVideoUrl] = useState(lesson.video?.source_url || "");
  const saveTimeoutRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => {
    setTitle(lesson.title);
    setVideoUrl(lesson.video?.source_url || "");
  }, [lesson.id]);

  const save = useCallback(
    async (updates: any) => {
      try {
        const { lesson: updated } = await updateLesson(courseId, lesson.id, updates);
        onUpdate(updated);
      } catch {
        toast.error("Failed to save");
      }
    },
    [courseId, lesson.id, onUpdate]
  );

  const debouncedSave = useCallback(
    (updates: any) => {
      if (saveTimeoutRef.current) clearTimeout(saveTimeoutRef.current);
      saveTimeoutRef.current = setTimeout(() => save(updates), 2000);
    },
    [save]
  );

  return (
    <div className="space-y-4 max-w-3xl">
      <div>
        <Label>Title</Label>
        <Input
          value={title}
          onChange={(e) => {
            setTitle(e.target.value);
            debouncedSave({ title: e.target.value });
          }}
          maxLength={200}
        />
      </div>

      {lesson.type === "video" && (
        <div className="space-y-4">
          <div>
            <Label>YouTube URL</Label>
            <Input
              value={videoUrl}
              onChange={(e) => {
                setVideoUrl(e.target.value);
                const pid = parseYouTubeUrl(e.target.value);
                if (pid) {
                  debouncedSave({
                    video: { provider: "youtube", provider_id: pid, source_url: e.target.value },
                  });
                }
              }}
              placeholder="https://www.youtube.com/watch?v=..."
            />
          </div>
          <div>
            <Label>Curator Notes</Label>
            <TiptapEditor
              content={lesson.video?.curator_notes}
              onChange={(doc) => debouncedSave({ video: { ...lesson.video, curator_notes: doc } })}
            />
          </div>
        </div>
      )}

      {lesson.type === "page" && (
        <div>
          <Label>Content</Label>
          <TiptapEditor
            content={lesson.page?.content}
            onChange={(doc) => debouncedSave({ page: { content: doc } })}
          />
        </div>
      )}

      {lesson.type === "quiz" && (
        <div>
          <Label>Questions (JSON)</Label>
          <Textarea
            defaultValue={JSON.stringify(lesson.quiz?.questions || [], null, 2)}
            rows={10}
            onChange={(e) => {
              try {
                const questions = JSON.parse(e.target.value);
                debouncedSave({ quiz: { questions } });
              } catch {}
            }}
          />
        </div>
      )}
    </div>
  );
}
