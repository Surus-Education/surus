"use client";

import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import type { TiptapDoc } from "@/lib/types";

export function TiptapRenderer({ content }: { content: TiptapDoc }) {
  const editor = useEditor({
    extensions: [StarterKit],
    content,
    editable: false,
    immediatelyRender: false,
  });

  return <EditorContent editor={editor} className="prose prose-sm max-w-none" />;
}
