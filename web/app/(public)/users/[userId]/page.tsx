import { getUser } from "@/lib/api/server/users";
import { notFound } from "next/navigation";
import { CourseCard } from "@/components/course/CourseCard";

export default async function UserProfilePage({
  params,
}: {
  params: Promise<{ userId: string }>;
}) {
  const { userId } = await params;

  let userData;
  try {
    userData = await getUser(userId);
  } catch {
    notFound();
  }

  const { user, courses = [] } = userData;

  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <div className="mb-8">
        <div className="flex items-center gap-4">
          {user.avatar_url && (
            <img
              src={user.avatar_url}
              alt={user.display_name}
              className="w-16 h-16 rounded-full"
            />
          )}
          <div>
            <h1 className="text-2xl font-bold">{user.display_name}</h1>
            {user.bio && <p className="text-muted-foreground mt-1">{user.bio}</p>}
          </div>
        </div>
      </div>

      <h2 className="text-lg font-semibold mb-4">Courses</h2>
      {courses.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          {courses.map((course: any) => (
            <CourseCard key={course.id} course={course} />
          ))}
        </div>
      ) : (
        <p className="text-muted-foreground">No public courses yet.</p>
      )}
    </div>
  );
}
