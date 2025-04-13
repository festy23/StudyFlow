CREATE TABLE users (
   id UUID PRIMARY KEY,
   role VARCHAR(16) NOT NULL,
   auth_provider VARCHAR(32) NOT NULL,
   status VARCHAR(16) NOT NULL DEFAULT 'active',
   first_name VARCHAR(64),
   last_name VARCHAR(64),
   timezone VARCHAR(64),
   created_at TIMESTAMP NOT NULL DEFAULT now(),
   edited_at TIMESTAMP NOT NULL DEFAULT now()
);
