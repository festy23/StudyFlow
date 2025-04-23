CREATE TABLE IF NOT EXISTS "receipts" (
  "id" uuid PRIMARY KEY,
  "lesson_id" uuid NOT NULL,
  "file_id" uuid NOT NULL,
  "is_verified" boolean NOT NULL DEFAULT false,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "edited_at" timestamp NOT NULL DEFAULT now()
);

COMMENT ON COLUMN "receipts"."id" IS 'UUIDv7';

COMMENT ON COLUMN "receipts"."lesson_id" IS 'Refers to schedule.lessons.id';

COMMENT ON COLUMN "receipts"."file_id" IS 'Refers to file_service.files.id';
