DO $$ BEGIN
  CREATE EXTENSION pgcrypto;
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE projects (
  id UUID NOT NULL PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMP
);
