import { AlertTriangle } from "lucide-react";

export function BrokenEmbedBanner() {
  return (
    <div className="flex items-center gap-2 text-sm text-yellow-800 bg-yellow-50 border border-yellow-200 rounded-md px-3 py-2">
      <AlertTriangle className="h-4 w-4" />
      <span>Some video embeds in this course may be broken or unavailable.</span>
    </div>
  );
}
