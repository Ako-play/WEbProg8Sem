ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT;

UPDATE users
SET username = 'user_' || REPLACE(id::text, '-', '')
WHERE username IS NULL OR TRIM(username) = '';

ALTER TABLE users ALTER COLUMN username SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS users_username_lower_uidx ON users (lower(trim(username)));
