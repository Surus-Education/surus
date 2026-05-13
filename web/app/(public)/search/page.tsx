"use client";

import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { getCourses } from "@/lib/api/courses";
import { CourseCard } from "@/components/course/CourseCard";
import Link from "next/link";
import { Suspense } from "react";

function SearchResults() {
  const searchParams = useSearchParams();
  const q = searchParams.get("q") || "";
  const tags = searchParams.get("tags") || "";

  const { data, isLoading } = useQuery({
    queryKey: ["courses", "search", q, tags],
    queryFn: () => getCourses({ q, tags }),
    enabled: q.length > 0 || tags.length > 0,
  });

  const courses = data?.data ?? [];

  return (
    <div className="mx-auto max-w-7xl px-4 py-8">
      <h1 className="text-2xl font-bold mb-6">
        {q ? `Search results for "${q}"` : "Search courses"}
      </h1>

      {isLoading && <p className="text-muted-foreground">Searching...</p>}

      {!isLoading && courses.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {courses.map((course) => (
            <CourseCard key={course.id} course={course} />
          ))}
        </div>
      )}

      {!isLoading && courses.length === 0 && q && (
        <div className="text-center py-16">
          <p className="text-muted-foreground mb-4">No results found.</p>
          <Link href="/courses/new" className="text-primary underline">
            Create a course
          </Link>
        </div>
      )}
    </div>
  );
}

export default function SearchPage() {
  return (
    <Suspense fallback={<div className="mx-auto max-w-7xl px-4 py-8">Loading...</div>}>
      <SearchResults />
    </Suspense>
  );
}
