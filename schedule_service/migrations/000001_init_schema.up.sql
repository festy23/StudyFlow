-- Слоты 
CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY,
    tutor_id UUID NOT NULL,
    starts_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_booked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    edited_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() 
);

CREATE INDEX idx_slots_tutor ON slots(tutor_id);
CREATE INDEX idx_slots_availability ON slots(tutor_id, is_booked);
CREATE INDEX idx_slots_starts_at ON slots(starts_at);

-- Уроки
CREATE TABLE IF NOT EXISTS lessons (
    id UUID PRIMARY KEY,
    slot_id UUID NOT NULL REFERENCES slots(id),
    student_id UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('booked', 'cancelled', 'completed')),
    is_paid BOOLEAN NOT NULL DEFAULT FALSE,
    connection_link TEXT,
    price_rub INTEGER,
    payment_info TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    edited_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    CONSTRAINT unique_slot_lesson UNIQUE (slot_id)
);

CREATE INDEX idx_lessons_student ON lessons(student_id);
CREATE INDEX idx_lessons_status ON lessons(status);
CREATE INDEX idx_lessons_paid ON lessons(is_paid);
CREATE INDEX idx_lessons_created ON lessons(created_at);
CREATE INDEX idx_lessons_student_status ON lessons(student_id, status);
CREATE INDEX idx_lessons_tutor ON lessons(slot_id) WHERE status = 'booked';
CREATE INDEX idx_lessons_price ON lessons(price_rub) WHERE price_rub IS NOT NULL;
CREATE INDEX idx_lessons_time_range ON lessons(slot_id, created_at);
CREATE INDEX idx_lessons_status_paid ON lessons(status, is_paid);
CREATE INDEX idx_slots_time_range ON slots(starts_at, ends_at) WHERE is_booked = false;
CREATE UNIQUE INDEX idx_slots_tutor_time_unique ON slots(tutor_id, starts_at, ends_at);
