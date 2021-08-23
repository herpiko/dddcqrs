DO $$ BEGIN
  CREATE EXTENSION pgcrypto;
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

CREATE TABLE articles (
  id SERIAL NOT NULL PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  author TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
