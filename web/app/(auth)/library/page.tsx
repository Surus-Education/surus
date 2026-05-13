"use client";

import { useQuery } from "@tanstack/react-query";
import { getLibrary } from "@/lib/api/users";
import { CourseCard } from "@/components/course/CourseCard";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./Tabs";

export default function LibraryPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["users", "me", "library"],
    queryFn: getLibrary,
  });

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <h1 className="text-2xl font-bold mb-6">Library</h1>

      {isLoading && <p className="text-muted-foreground">Loading...</p>}

      {data && (
        <div className="space-y-8">
          <section>
            <h2 className="text-lg font-semibold mb-4">Saved ({(data.saved ?? []).length})</h2>
            {(data.saved ?? []).length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {(data.saved ?? []).map((course) => (
                  <CourseCard key={course.id} course={course} />
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground">No saved courses.</p>
            )}
          </section>

          <section>
            <h2 className="text-lg font-semibold mb-4">Created ({(data.created ?? []).length})</h2>
            {(data.created ?? []).length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {(data.created ?? []).map((course) => (
                  <CourseCard key={course.id} course={course} />
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground">No created courses.</p>
            )}
          </section>
        </div>
      )}
    </div>
  );
}
