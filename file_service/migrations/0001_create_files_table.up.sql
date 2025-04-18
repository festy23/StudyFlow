CREATE TABLE files (
    id UUID PRIMARY KEY,
    extension VARCHAR(32) NOT NULL,
    uploaded_by UUID NOT NULL,
    filename TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);