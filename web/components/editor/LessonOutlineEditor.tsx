"use client";

import { useState } from "react";
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
  useSortable,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, Plus, Trash2, Video, FileText, HelpCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import type { Lesson } from "@/lib/types";
import { cn } from "@/lib/utils";

function SortableLesson({
  lesson,
  isSelected,
  onSelect,
  onDelete,
}: {
  lesson: Lesson;
  isSelected: boolean;
  onSelect: () => void;
  onDelete: () => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: lesson.id,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const Icon = lesson.type === "video" ? Video : lesson.type === "page" ? FileText : HelpCircle;

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm border cursor-pointer",
        isSelected && "border-primary bg-primary/5",
        !isSelected && "border-transparent hover:bg-accent",
        isDragging && "opacity-50"
      )}
      onClick={onSelect}
    >
      <button {...attributes} {...listeners} className="cursor-grab touch-none">
        <GripVertical className="h-4 w-4 text-muted-foreground" />
      </button>
      <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
      <span className="truncate flex-1">{lesson.title || "Untitled"}</span>
      <button
        onClick={(e) => {
          e.stopPropagation();
          onDelete();
        }}
        className="text-muted-foreground hover:text-destructive"
      >
        <Trash2 className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}

export function LessonOutlineEditor({
  lessons,
  selectedId,
  onSelect,
  onReorder,
  onAdd,
  onDelete,
}: {
  lessons: Lesson[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onReorder: (lessons: Lesson[]) => void;
  onAdd: (type: "video" | "page" | "quiz") => void;
  onDelete: (id: string) => void;
}) {
  const [addMenuOpen, setAddMenuOpen] = useState(false);
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }));

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (over && active.id !== over.id) {
      const oldIndex = lessons.findIndex((l) => l.id === active.id);
      const newIndex = lessons.findIndex((l) => l.id === over.id);
      onReorder(arrayMove(lessons, oldIndex, newIndex));
    }
  }

  return (
    <div className="space-y-2">
      <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={lessons.map((l) => l.id)} strategy={verticalListSortingStrategy}>
          {lessons.map((lesson) => (
            <SortableLesson
              key={lesson.id}
              lesson={lesson}
              isSelected={selectedId === lesson.id}
              onSelect={() => onSelect(lesson.id)}
              onDelete={() => onDelete(lesson.id)}
            />
          ))}
        </SortableContext>
      </DndContext>

      <div className="relative">
        <Button
          variant="outline"
          size="sm"
          className="w-full"
          onClick={() => setAddMenuOpen(!addMenuOpen)}
        >
          <Plus className="h-4 w-4 mr-1" /> Add lesson
        </Button>
        {addMenuOpen && (
          <div className="absolute top-full left-0 right-0 mt-1 bg-popover border rounded-md shadow-md z-10 py-1">
            {[
              { type: "video" as const, icon: Video, label: "Video" },
              { type: "page" as const, icon: FileText, label: "Page" },
              { type: "quiz" as const, icon: HelpCircle, label: "Quiz" },
            ].map(({ type, icon: Icon, label }) => (
              <button
                key={type}
                className="flex items-center gap-2 w-full px-3 py-1.5 text-sm hover:bg-accent"
                onClick={() => {
                  onAdd(type);
                  setAddMenuOpen(false);
                }}
              >
                <Icon className="h-4 w-4" /> {label}
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
