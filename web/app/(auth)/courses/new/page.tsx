"use client";

import { useState, useCallback, useEffect, useRef } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { LessonOutlineEditor } from "@/components/editor/LessonOutlineEditor";
import { TiptapEditor } from "@/components/editor/TiptapEditor";
import { createCourse } from "@/lib/api/courses";
import { createLesson, updateLesson, deleteLesson, reorderLessons } from "@/lib/api/lessons";
import type { Lesson, TiptapDoc, CourseInput } from "@/lib/types";
import { toast } from "sonner";

export default function NewCoursePage() {
  const router = useRouter();
  const queryClient = useQueryClient();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState("");
  const [visibility, setVisibility] = useState<"public" | "unlisted" | "private">("private");
  const [courseId, setCourseId] = useState<string | null>(null);
  const [lessons, setLessons] = useState<Lesson[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [videoUrl, setVideoUrl] = useState("");
  const [pageContent, setPageContent] = useState<TiptapDoc | null>(null);
  const [curatorNotes, setCuratorNotes] = useState<TiptapDoc | null>(null);

  const selectedLesson = lessons.find((l) => l.id === selectedId);

  const createCourseMutation = useMutation({
    mutationFn: (input: CourseInput) => createCourse(input),
    onSuccess: (data) => {
      setCourseId(data.course.id);
      toast.success("Course created");
    },
    onError: () => toast.error("Failed to create course"),
  });

  const handleCreateCourse = () => {
    if (!title.trim()) {
      toast.error("Title is required");
      return;
    }
    createCourseMutation.mutate({ title, description, tags, visibility });
  };

  const handleAddLesson = async (type: "video" | "page" | "quiz") => {
    if (!courseId) return;
    try {
      const input: any = {
        type,
        title: `New ${type} lesson`,
        position: lessons.length,
      };
      if (type === "page") {
        input.page = { content: { type: "doc", content: [{ type: "paragraph" }] } };
      }
      if (type === "quiz") {
        input.quiz = { questions: [] };
      }
      if (type === "video") {
        input.video = { provider: "youtube", provider_id: "placeholder", source_url: "https://youtube.com" };
      }

      const { lesson } = await createLesson(courseId, input);
      setLessons([...lessons, lesson]);
      setSelectedId(lesson.id);
    } catch {
      toast.error("Failed to add lesson");
    }
  };

  const handleDeleteLesson = async (id: string) => {
    if (!courseId) return;
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
    if (courseId) {
      const moves = newLessons.map((l, i) => ({
        lesson_id: l.id,
        new_parent_id: l.parent_id,
        new_position: i,
      }));
      reorderLessons(courseId, moves).catch(() => toast.error("Failed to save order"));
    }
  };

  const handleTagKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if ((e.key === "Enter" || e.key === ",") && tagInput.trim()) {
      e.preventDefault();
      const newTag = tagInput.trim().toLowerCase();
      if (!tags.includes(newTag) && tags.length < 10) {
        setTags([...tags, newTag]);
      }
      setTagInput("");
    }
  };

  const removeTag = (tag: string) => {
    setTags(tags.filter((t) => t !== tag));
  };

  if (!courseId) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-8">
        <h1 className="text-2xl font-bold mb-6">Create a new course</h1>
        <div className="space-y-4">
          <div>
            <Label>Title</Label>
            <Input value={title} onChange={(e) => setTitle(e.target.value)} maxLength={200} />
          </div>
          <div>
            <Label>Description</Label>
            <Textarea value={description} onChange={(e) => setDescription(e.target.value)} maxLength={5000} />
          </div>
          <div>
            <Label>Tags</Label>
            <div className="flex flex-wrap gap-1 mb-2">
              {tags.map((tag) => (
                <span key={tag} className="inline-flex items-center gap-1 rounded-md bg-secondary px-2 py-0.5 text-xs">
                  {tag}
                  <button onClick={() => removeTag(tag)} className="hover:text-destructive">&times;</button>
                </span>
              ))}
            </div>
            <Input
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={handleTagKeyDown}
              placeholder="Type a tag and press Enter"
            />
          </div>
          <div>
            <Label>Visibility</Label>
            <Select value={visibility} onValueChange={(v: any) => setVisibility(v)}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="private">Private</SelectItem>
                <SelectItem value="unlisted">Unlisted</SelectItem>
                <SelectItem value="public">Public</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Button onClick={handleCreateCourse} disabled={createCourseMutation.isPending}>
            {createCourseMutation.isPending ? "Creating..." : "Create course"}
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-[calc(100vh-3.5rem)]">
      <div className="w-72 border-r p-4 overflow-y-auto">
        <div className="mb-4">
          <Input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Course title" className="font-semibold" />
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
                    video: {
                      provider: "youtube",
                      provider_id: pid,
                      source_url: e.target.value,
                    },
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
              onChange={(doc) => {
                debouncedSave({ video: { ...lesson.video, curator_notes: doc } });
              }}
            />
          </div>
        </div>
      )}

      {lesson.type === "page" && (
        <div>
          <Label>Content</Label>
          <TiptapEditor
            content={lesson.page?.content}
            onChange={(doc) => {
              debouncedSave({ page: { content: doc } });
            }}
          />
        </div>
      )}

      {lesson.type === "quiz" && (
        <div>
          <Label>Questions</Label>
          <p className="text-sm text-muted-foreground">Quiz editor — add questions using the JSON format for now.</p>
          <Textarea
            defaultValue={JSON.stringify(lesson.quiz?.questions || [], null, 2)}
            rows={10}
            onChange={(e) => {
              try {
                const questions = JSON.parse(e.target.value);
                debouncedSave({ quiz: { questions } });
              } catch {
                // Invalid JSON, don't save
              }
            }}
          />
        </div>
      )}
    </div>
  );
}

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
