CREATE TABLE receipts (
    id UUID PRIMARY KEY,
    lesson_id UUID NOT NULL,
    file_id UUID NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL,
    edited_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_receipts_lesson_id ON receipts (lesson_id);