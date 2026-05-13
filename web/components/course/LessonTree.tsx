"use client";

import Link from "next/link";
import { CheckCircle, Circle, FileText, Video, HelpCircle, AlertTriangle } from "lucide-react";
import type { Lesson } from "@/lib/types";
import { cn } from "@/lib/utils";

function buildTree(lessons: Lesson[]) {
  const roots: (Lesson & { children: Lesson[] })[] = [];
  const map = new Map<string, Lesson & { children: Lesson[] }>();

  for (const l of lessons) {
    map.set(l.id, { ...l, children: [] });
  }

  for (const l of lessons) {
    const node = map.get(l.id)!;
    if (l.parent_id && map.has(l.parent_id)) {
      map.get(l.parent_id)!.children.push(node);
    } else {
      roots.push(node);
    }
  }

  return roots;
}

const typeIcon = {
  video: Video,
  page: FileText,
  quiz: HelpCircle,
};

function LessonNode({
  lesson,
  courseId,
  completedLessonIds,
  depth = 0,
}: {
  lesson: Lesson & { children: Lesson[] };
  courseId: string;
  completedLessonIds: Set<string>;
  depth?: number;
}) {
  const Icon = typeIcon[lesson.type];
  const isComplete = completedLessonIds.has(lesson.id);

  return (
    <div>
      <Link
        href={`/courses/${courseId}/lessons/${lesson.id}`}
        className={cn(
          "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent transition-colors",
          depth > 0 && "ml-4"
        )}
      >
        {isComplete ? (
          <CheckCircle className="h-4 w-4 text-green-600 shrink-0" />
        ) : (
          <Circle className="h-4 w-4 text-muted-foreground shrink-0" />
        )}
        <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="truncate">{lesson.title}</span>
        {lesson.embed_broken && (
          <AlertTriangle className="h-3.5 w-3.5 text-yellow-500 shrink-0" />
        )}
      </Link>
      {lesson.children.map((child: any) => (
        <LessonNode
          key={child.id}
          lesson={child}
          courseId={courseId}
          completedLessonIds={completedLessonIds}
          depth={depth + 1}
        />
      ))}
    </div>
  );
}

export function LessonTree({
  lessons,
  courseId,
  completedLessonIds = new Set(),
}: {
  lessons: Lesson[];
  courseId: string;
  completedLessonIds?: Set<string>;
}) {
  const tree = buildTree(lessons);

  if (tree.length === 0) {
    return <p className="text-sm text-muted-foreground py-4">No lessons yet.</p>;
  }

  return (
    <div className="space-y-0.5">
      {tree.map((node) => (
        <LessonNode
          key={node.id}
          lesson={node}
          courseId={courseId}
          completedLessonIds={completedLessonIds}
        />
      ))}
    </div>
  );
}
