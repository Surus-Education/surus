"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import type { TiptapDoc } from "@/lib/types";

export function TiptapEditor({
  content,
  onChange,
  placeholder = "Start writing...",
}: {
  content?: TiptapDoc | null;
  onChange: (doc: TiptapDoc) => void;
  placeholder?: string;
}) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder }),
    ],
    content: content || undefined,
    onUpdate: ({ editor }) => {
      onChange(editor.getJSON() as TiptapDoc);
    },
    immediatelyRender: false,
  });

  return (
    <div className="border rounded-md">
      <EditorContent editor={editor} className="prose prose-sm max-w-none p-3 min-h-[120px]" />
    </div>
  );
}
