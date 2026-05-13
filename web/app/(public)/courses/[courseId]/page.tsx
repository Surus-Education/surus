import { getCourse, getLessons } from "@/lib/api/server/courses";
import { notFound } from "next/navigation";
import { LessonTree } from "@/components/course/LessonTree";
import { ForkBanner } from "@/components/course/ForkBanner";
import { BrokenEmbedBanner } from "@/components/course/BrokenEmbedBanner";
import { Badge } from "@/components/ui/badge";
import { CourseActions } from "./CourseActions";

export default async function CoursePage({
  params,
}: {
  params: Promise<{ courseId: string }>;
}) {
  const { courseId } = await params;

  let course, lessons;
  try {
    const [courseData, lessonsData] = await Promise.all([
      getCourse(courseId),
      getLessons(courseId),
    ]);
    course = courseData.course;
    lessons = lessonsData.lessons;
  } catch {
    notFound();
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <div className="space-y-4">
        {course.forked_from_id && (
          <ForkBanner forkedFromId={course.forked_from_id} forkedAt={course.forked_at!} />
        )}

        {course.embed_broken && <BrokenEmbedBanner />}

        <div className="flex gap-6">
          <div className="flex-1 space-y-4">
            <h1 className="text-3xl font-bold">{course.title}</h1>

            {course.description && (
              <p className="text-muted-foreground">{course.description}</p>
            )}

            {course.tags.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {course.tags.map((tag) => (
                  <Badge key={tag} variant="secondary">
                    {tag}
                  </Badge>
                ))}
              </div>
            )}

            <CourseActions courseId={courseId} />
          </div>

          {course.thumbnail_url && (
            <div className="w-64 shrink-0 hidden md:block">
              <img
                src={course.thumbnail_url}
                alt={course.title}
                className="w-full rounded-lg"
              />
            </div>
          )}
        </div>

        <div className="border-t pt-6">
          <h2 className="text-lg font-semibold mb-3">Lessons</h2>
          <LessonTree lessons={lessons} courseId={courseId} />
        </div>
      </div>
    </div>
  );
}
