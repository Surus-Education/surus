"use client";

import { useAuth } from "@/hooks/useAuth";
import { useQuery } from "@tanstack/react-query";
import { getLibrary } from "@/lib/api/users";
import { CourseCard } from "@/components/course/CourseCard";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { Plus } from "lucide-react";

export default function DashboardPage() {
  const { user } = useAuth();
  const { data, isLoading } = useQuery({
    queryKey: ["users", "me", "library"],
    queryFn: getLibrary,
  });

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold">Dashboard</h1>
          {user && <p className="text-muted-foreground">Welcome back, {user.display_name}</p>}
        </div>
        <Button asChild>
          <Link href="/courses/new">
            <Plus className="h-4 w-4 mr-1" />
            Create course
          </Link>
        </Button>
      </div>

      {isLoading && <p className="text-muted-foreground">Loading...</p>}

      {data && (
        <div className="space-y-8">
          <section>
            <h2 className="text-lg font-semibold mb-4">Your courses</h2>
            {(data.created ?? []).length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {(data.created ?? []).map((course) => (
                  <CourseCard key={course.id} course={course} />
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground">
                You haven&apos;t created any courses yet.{" "}
                <Link href="/courses/new" className="text-primary underline">
                  Create one
                </Link>
              </p>
            )}
          </section>

          <section>
            <h2 className="text-lg font-semibold mb-4">Saved courses</h2>
            {(data.saved ?? []).length > 0 ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                {(data.saved ?? []).map((course) => (
                  <CourseCard key={course.id} course={course} />
                ))}
              </div>
            ) : (
              <p className="text-muted-foreground">
                You haven&apos;t saved any courses yet. Browse the{" "}
                <Link href="/" className="text-primary underline">
                  course library
                </Link>
              </p>
            )}
          </section>
        </div>
      )}
    </div>
  );
}
