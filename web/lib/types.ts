export interface User {
  id: string;
  email: string;
  display_name: string;
  bio: string | null;
  avatar_url: string | null;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
}

export interface PublicUser {
  id: string;
  display_name: string;
  bio: string | null;
  avatar_url: string | null;
  created_at: string;
}

export interface Course {
  id: string;
  owner_id: string;
  title: string;
  description: string;
  tags: string[];
  thumbnail_url: string | null;
  visibility: "public" | "unlisted" | "private";
  forked_from_id: string | null;
  forked_at: string | null;
  embed_broken: boolean;
  created_at: string;
  updated_at: string;
}

export interface CourseInput {
  title: string;
  description?: string;
  tags?: string[];
  thumbnail_url?: string;
  visibility: "public" | "unlisted" | "private";
}

export interface Lesson {
  id: string;
  course_id: string;
  parent_id: string | null;
  position: number;
  type: "video" | "page" | "quiz";
  title: string;
  embed_broken: boolean;
  created_at: string;
  updated_at: string;
  video?: VideoDetail;
  page?: PageDetail;
  quiz?: QuizDetail;
}

export interface VideoDetail {
  provider: "youtube";
  provider_id: string;
  start_seconds: number | null;
  end_seconds: number | null;
  curator_notes: TiptapDoc | null;
  source_url: string;
}

export interface PageDetail {
  content: TiptapDoc;
}

export interface QuizDetail {
  questions: QuizQuestion[];
}

export interface TiptapDoc {
  type: "doc";
  content: any[];
}

export interface QuizQuestion {
  id: string;
  prompt: TiptapDoc;
  type: "multiple_choice" | "short_answer";
  options?: QuizOption[];
  explanation: TiptapDoc;
}

export interface QuizOption {
  id: string;
  text: string;
  correct: boolean;
}

export interface LessonInput {
  parent_id?: string | null;
  position?: number;
  type: "video" | "page" | "quiz";
  title: string;
  video?: {
    provider: "youtube";
    provider_id: string;
    start_seconds?: number;
    end_seconds?: number;
    curator_notes?: TiptapDoc | null;
    source_url: string;
  };
  page?: {
    content: TiptapDoc;
  };
  quiz?: {
    questions: QuizQuestion[];
  };
}

export interface ReorderMove {
  lesson_id: string;
  new_parent_id: string | null;
  new_position: number;
}

export interface QuizAttemptResult {
  attempt_id?: string;
  score?: number;
  question_results: { correct: boolean; explanation?: TiptapDoc }[];
}

export interface Report {
  id: string;
  reporter_id: string;
  target_type: "course" | "lesson";
  target_id: string;
  category: "incorrect" | "harmful" | "copyright" | "other";
  body: string | null;
  status: "open" | "reviewed" | "actioned" | "dismissed";
  reviewed_by: string | null;
  reviewed_at: string | null;
  created_at: string;
}

export interface ReportInput {
  category: "incorrect" | "harmful" | "copyright" | "other";
  body?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  next_cursor: string | null;
}

export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
  };
}
