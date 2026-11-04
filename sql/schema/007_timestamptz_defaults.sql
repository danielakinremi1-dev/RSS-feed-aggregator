-- +goose Up
-- TIMESTAMP drops the UTC offset Go sends, so instants from different zones
-- were compared as wall-clock times. TIMESTAMPTZ stores real instants.
-- Existing values are interpreted in the session time zone.
ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMPTZ, ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ, ALTER COLUMN updated_at SET DEFAULT now();

ALTER TABLE feeds
    ALTER COLUMN created_at TYPE TIMESTAMPTZ, ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ, ALTER COLUMN updated_at SET DEFAULT now(),
    ALTER COLUMN last_fetched_at TYPE TIMESTAMPTZ;

ALTER TABLE feed_follows
    ALTER COLUMN created_at TYPE TIMESTAMPTZ, ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ, ALTER COLUMN updated_at SET DEFAULT now();

ALTER TABLE posts
    ALTER COLUMN created_at TYPE TIMESTAMPTZ, ALTER COLUMN created_at SET DEFAULT now(),
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ, ALTER COLUMN updated_at SET DEFAULT now(),
    ALTER COLUMN published_at TYPE TIMESTAMPTZ;

-- +goose Down
ALTER TABLE posts
    ALTER COLUMN created_at DROP DEFAULT, ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at DROP DEFAULT, ALTER COLUMN updated_at TYPE TIMESTAMP,
    ALTER COLUMN published_at TYPE TIMESTAMP;

ALTER TABLE feed_follows
    ALTER COLUMN created_at DROP DEFAULT, ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at DROP DEFAULT, ALTER COLUMN updated_at TYPE TIMESTAMP;

ALTER TABLE feeds
    ALTER COLUMN created_at DROP DEFAULT, ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at DROP DEFAULT, ALTER COLUMN updated_at TYPE TIMESTAMP,
    ALTER COLUMN last_fetched_at TYPE TIMESTAMP;

ALTER TABLE users
    ALTER COLUMN created_at DROP DEFAULT, ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at DROP DEFAULT, ALTER COLUMN updated_at TYPE TIMESTAMP;
