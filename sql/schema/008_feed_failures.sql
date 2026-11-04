-- +goose Up
ALTER TABLE feeds
    ADD COLUMN last_error TEXT,
    ADD COLUMN consecutive_failures INT NOT NULL DEFAULT 0,
    -- Earliest time the feed may be fetched again; NULL means "due now".
    ADD COLUMN next_fetch_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE feeds
    DROP COLUMN next_fetch_at,
    DROP COLUMN consecutive_failures,
    DROP COLUMN last_error;
