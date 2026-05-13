import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import type { Course } from "@/lib/types";

export function CourseCard({ course }: { course: Course }) {
  return (
    <Link href={`/courses/${course.id}`}>
      <Card className="overflow-hidden hover:shadow-md transition-shadow h-full">
        {course.thumbnail_url && (
          <div className="aspect-video bg-muted">
            <img
              src={course.thumbnail_url}
              alt={course.title}
              className="w-full h-full object-cover"
            />
          </div>
        )}
        {!course.thumbnail_url && (
          <div className="aspect-video bg-muted flex items-center justify-center text-muted-foreground text-sm">
            No thumbnail
          </div>
        )}
        <CardContent className="p-4">
          <h3 className="font-semibold line-clamp-2 mb-1">{course.title}</h3>
          {course.description && (
            <p className="text-sm text-muted-foreground line-clamp-2 mb-2">
              {course.description}
            </p>
          )}
          {course.tags.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {course.tags.slice(0, 3).map((tag) => (
                <Badge key={tag} variant="secondary" className="text-xs">
                  {tag}
                </Badge>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </Link>
  );
}
