CREATE TABLE IF NOT EXISTS lesson_courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category VARCHAR(24) NOT NULL
        CHECK (category IN ('opening', 'middlegame', 'endgame', 'tactics', 'strategy')),
    cover_image_url TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(16) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS lesson_courses_published_order_idx
    ON lesson_courses (sort_order, title) WHERE status = 'published';

CREATE TABLE IF NOT EXISTS lesson_chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES lesson_courses(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    content JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(content) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (course_id, sort_order)
);

CREATE INDEX IF NOT EXISTS lesson_chapters_course_order_idx
    ON lesson_chapters (course_id, sort_order);

CREATE TABLE IF NOT EXISTS user_lesson_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chapter_id UUID NOT NULL REFERENCES lesson_chapters(id) ON DELETE CASCADE,
    progress_percent SMALLINT NOT NULL DEFAULT 0 CHECK (progress_percent BETWEEN 0 AND 100),
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    last_position JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(last_position) = 'object'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, chapter_id),
    CONSTRAINT lesson_completed_progress CHECK (NOT completed OR progress_percent = 100)
);

CREATE INDEX IF NOT EXISTS user_lesson_progress_user_updated_idx
    ON user_lesson_progress (user_id, updated_at DESC);
