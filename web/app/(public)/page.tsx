import { getCourses } from "@/lib/api/server/courses";
import { CourseCard } from "@/components/course/CourseCard";

export default async function HomePage() {
  let courses: any[] = [];
  try {
    const data = await getCourses({ limit: 24 });
    courses = data.data ?? [];
  } catch {
    courses = [];
  }

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Explore courses</h1>
        <p className="text-muted-foreground">
          Discover community-curated courses built from YouTube videos, written pages, and quizzes.
        </p>
      </div>

      {courses.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {courses.map((course) => (
            <CourseCard key={course.id} course={course} />
          ))}
        </div>
      ) : (
        <div className="text-center py-16">
          <p className="text-muted-foreground mb-4">No courses yet. Be the first to create one!</p>
        </div>
      )}
    </div>
  );
}
