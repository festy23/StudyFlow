CREATE TABLE telegram_accounts (
   id UUID PRIMARY KEY,
   user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
   telegram_id BIGINT NOT NULL UNIQUE,
   username VARCHAR(32),
   created_at TIMESTAMP NOT NULL DEFAULT now()
);
