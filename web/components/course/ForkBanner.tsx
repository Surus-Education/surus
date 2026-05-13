import Link from "next/link";
import { GitFork } from "lucide-react";

export function ForkBanner({
  forkedFromId,
  forkedAt,
}: {
  forkedFromId: string;
  forkedAt: string;
}) {
  return (
    <div className="flex items-center gap-2 text-sm text-muted-foreground bg-muted/50 rounded-md px-3 py-2">
      <GitFork className="h-4 w-4" />
      <span>
        Forked from{" "}
        <Link href={`/courses/${forkedFromId}`} className="underline hover:text-foreground">
          original course
        </Link>
        {forkedAt && (
          <> on {new Date(forkedAt).toLocaleDateString()}</>
        )}
      </span>
    </div>
  );
}
