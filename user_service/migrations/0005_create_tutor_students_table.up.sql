CREATE TABLE tutor_students (
    id UUID PRIMARY KEY,
    tutor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_price_rub INTEGER,
    lesson_connection_link TEXT,
    status VARCHAR(16) NOT NULL DEFAULT 'invited',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    edited_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (tutor_id, student_id)
);
