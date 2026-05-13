-- +goose Up

-- Enumerations
CREATE TYPE lesson_type AS ENUM ('video', 'page', 'quiz');
CREATE TYPE visibility AS ENUM ('public', 'unlisted', 'private');
CREATE TYPE video_provider AS ENUM ('youtube');
CREATE TYPE question_type AS ENUM ('multiple_choice', 'short_answer');
CREATE TYPE report_target_type AS ENUM ('course', 'lesson');
CREATE TYPE report_category AS ENUM ('incorrect', 'harmful', 'copyright', 'other');
CREATE TYPE report_status AS ENUM ('open', 'reviewed', 'actioned', 'dismissed');

-- Users
CREATE TABLE users (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email                 TEXT UNIQUE NOT NULL,
  display_name          TEXT NOT NULL,
  bio                   TEXT,
  avatar_url            TEXT,
  is_admin              BOOLEAN NOT NULL DEFAULT false,
  deleted_at            TIMESTAMPTZ,
  deletion_scheduled_at TIMESTAMPTZ,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- System "deleted user" account for orphaned public courses
INSERT INTO users (id, email, display_name, is_admin)
VALUES ('00000000-0000-0000-0000-000000000001', 'deleted@system.local', 'Deleted user', false);

-- OAuth accounts
CREATE TABLE oauth_accounts (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  provider    TEXT NOT NULL,
  provider_id TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (provider, provider_id)
);

-- Magic link tokens
CREATE TABLE magic_link_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email       TEXT NOT NULL,
  token_hash  TEXT NOT NULL UNIQUE,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_magic_link_tokens_email ON magic_link_tokens(email);

-- Refresh tokens
CREATE TABLE refresh_tokens (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT NOT NULL UNIQUE,
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Courses
CREATE TABLE courses (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id        UUID NOT NULL REFERENCES users(id),
  title           TEXT NOT NULL,
  description     TEXT NOT NULL DEFAULT '',
  tags            TEXT[] NOT NULL DEFAULT '{}',
  thumbnail_url   TEXT,
  visibility      visibility NOT NULL DEFAULT 'private',
  forked_from_id  UUID REFERENCES courses(id) ON DELETE SET NULL,
  forked_at       TIMESTAMPTZ,
  embed_broken    BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  search_vector   TSVECTOR
);
CREATE INDEX idx_courses_owner ON courses(owner_id);
CREATE INDEX idx_courses_visibility ON courses(visibility);
CREATE INDEX idx_courses_search ON courses USING GIN(search_vector);
CREATE INDEX idx_courses_tags ON courses USING GIN(tags);

-- Full-text search trigger
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION courses_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(array_to_string(NEW.tags, ' '), '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER courses_search_vector_trigger
  BEFORE INSERT OR UPDATE ON courses
  FOR EACH ROW EXECUTE FUNCTION courses_search_vector_update();

-- Lessons
CREATE TABLE lessons (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id     UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  parent_id     UUID REFERENCES lessons(id) ON DELETE CASCADE,
  position      INTEGER NOT NULL DEFAULT 0,
  type          lesson_type NOT NULL,
  title         TEXT NOT NULL,
  embed_broken  BOOLEAN NOT NULL DEFAULT false,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_lessons_course ON lessons(course_id);
CREATE INDEX idx_lessons_parent ON lessons(parent_id);
CREATE UNIQUE INDEX idx_lessons_position
  ON lessons (course_id, COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'::uuid), position);

-- Video lessons
CREATE TABLE video_lessons (
  lesson_id       UUID PRIMARY KEY REFERENCES lessons(id) ON DELETE CASCADE,
  provider        video_provider NOT NULL DEFAULT 'youtube',
  provider_id     TEXT NOT NULL,
  start_seconds   INTEGER,
  end_seconds     INTEGER,
  curator_notes   JSONB,
  source_url      TEXT NOT NULL
);

-- Page lessons
CREATE TABLE page_lessons (
  lesson_id   UUID PRIMARY KEY REFERENCES lessons(id) ON DELETE CASCADE,
  content     JSONB NOT NULL
);

-- Quiz lessons
CREATE TABLE quiz_lessons (
  lesson_id   UUID PRIMARY KEY REFERENCES lessons(id) ON DELETE CASCADE,
  questions   JSONB NOT NULL
);

-- Quiz attempts
CREATE TABLE quiz_attempts (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  lesson_id   UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
  answers     JSONB NOT NULL,
  score       NUMERIC(5,2),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_quiz_attempts_user ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_lesson ON quiz_attempts(lesson_id);

-- Saves (bookmarks)
CREATE TABLE saves (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  course_id   UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, course_id)
);

-- Completions
CREATE TABLE completions (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  lesson_id    UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
  completed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (user_id, lesson_id)
);

-- Reports
CREATE TABLE reports (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  reporter_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  target_type   report_target_type NOT NULL,
  target_id     UUID NOT NULL,
  category      report_category NOT NULL,
  body          TEXT,
  status        report_status NOT NULL DEFAULT 'open',
  reviewed_by   UUID REFERENCES users(id),
  reviewed_at   TIMESTAMPTZ,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_target ON reports(target_type, target_id);

-- +goose Down
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS completions;
DROP TABLE IF EXISTS saves;
DROP TABLE IF EXISTS quiz_attempts;
DROP TABLE IF EXISTS quiz_lessons;
DROP TABLE IF EXISTS page_lessons;
DROP TABLE IF EXISTS video_lessons;
DROP TABLE IF EXISTS lessons;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS magic_link_tokens;
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS report_status;
DROP TYPE IF EXISTS report_category;
DROP TYPE IF EXISTS report_target_type;
DROP TYPE IF EXISTS question_type;
DROP TYPE IF EXISTS video_provider;
DROP TYPE IF EXISTS visibility;
DROP TYPE IF EXISTS lesson_type;
DROP FUNCTION IF EXISTS courses_search_vector_update;
