import { getLesson, getLessons, getCourse } from "@/lib/api/server/courses";
import { notFound } from "next/navigation";
import { VideoLesson } from "@/components/lesson/VideoLesson";
import { PageLesson } from "@/components/lesson/PageLesson";
import { QuizLesson } from "@/components/lesson/QuizLesson";
import { LessonNavigation } from "./LessonNavigation";
import Link from "next/link";
import { ChevronRight } from "lucide-react";

export default async function LessonPage({
  params,
}: {
  params: Promise<{ courseId: string; lessonId: string }>;
}) {
  const { courseId, lessonId } = await params;

  let course, lesson, lessons;
  try {
    const [courseData, lessonData, lessonsData] = await Promise.all([
      getCourse(courseId),
      getLesson(courseId, lessonId),
      getLessons(courseId),
    ]);
    course = courseData.course;
    lesson = lessonData.lesson;
    lessons = lessonsData.lessons;
  } catch {
    notFound();
  }

  const flatLessons = lessons.sort((a, b) => a.position - b.position);
  const currentIndex = flatLessons.findIndex((l) => l.id === lessonId);
  const prevLesson = currentIndex > 0 ? flatLessons[currentIndex - 1] : null;
  const nextLesson = currentIndex < flatLessons.length - 1 ? flatLessons[currentIndex + 1] : null;

  const parent = lesson.parent_id
    ? lessons.find((l) => l.id === lesson.parent_id)
    : null;

  return (
    <div className="mx-auto max-w-4xl px-4 py-8">
      <nav className="flex items-center gap-1 text-sm text-muted-foreground mb-6">
        <Link href={`/courses/${courseId}`} className="hover:text-foreground">
          {course.title}
        </Link>
        {parent && (
          <>
            <ChevronRight className="h-3.5 w-3.5" />
            <Link
              href={`/courses/${courseId}/lessons/${parent.id}`}
              className="hover:text-foreground"
            >
              {parent.title}
            </Link>
          </>
        )}
        <ChevronRight className="h-3.5 w-3.5" />
        <span className="text-foreground">{lesson.title}</span>
      </nav>

      <h1 className="text-2xl font-bold mb-6">{lesson.title}</h1>

      <div className="mb-8">
        {lesson.type === "video" && lesson.video && (
          <VideoLesson video={lesson.video} />
        )}
        {lesson.type === "page" && lesson.page && (
          <PageLesson content={lesson.page.content} />
        )}
        {lesson.type === "quiz" && lesson.quiz && (
          <QuizLesson
            courseId={courseId}
            lessonId={lessonId}
            questions={lesson.quiz.questions}
          />
        )}
      </div>

      <LessonNavigation
        courseId={courseId}
        lessonId={lessonId}
        prevLesson={prevLesson}
        nextLesson={nextLesson}
      />
    </div>
  );
}
