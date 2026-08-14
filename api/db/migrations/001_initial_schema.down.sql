-- 001_initial_schema.down.sql
--
-- ROLLBACK. Every statement here is destructive: running this file drops all
-- application tables and every row in them.
--
-- Split out of 001_initial_schema.sql so that nothing destructive lives in a
-- file people are told to paste into a SQL console. Do not run this against
-- production. Do not run it to "reset" a database you care about.

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
