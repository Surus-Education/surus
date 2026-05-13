import type { TiptapDoc } from "@/lib/types";
import { TiptapRenderer } from "@/components/editor/TiptapRenderer";

export function PageLesson({ content }: { content: TiptapDoc }) {
  return (
    <div className="prose prose-sm max-w-none">
      <TiptapRenderer content={content} />
    </div>
  );
}
