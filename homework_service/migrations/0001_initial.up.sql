CREATE TABLE assignments (
    id UUID PRIMARY KEY,
    tutor_id UUID NOT NULL,
    student_id UUID NOT NULL,
    title TEXT,
    description TEXT,
    file_id UUID,
    due_date TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE submissions (
    id UUID PRIMARY KEY,
    assignment_id UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    file_id UUID,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMP NOT NULL DEFAULT NOW()  
);

CREATE TABLE feedbacks (
    id UUID PRIMARY KEY,
    submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    file_id UUID,
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    edited_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assignments_tutor_id ON assignments(tutor_id);
CREATE INDEX idx_assignments_student_id ON assignments(student_id);
CREATE INDEX idx_submissions_assignment_id ON submissions(assignment_id);
CREATE INDEX idx_feedbacks_submission_id ON feedbacks(submission_id);