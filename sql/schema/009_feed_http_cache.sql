-- +goose Up
-- HTTP validators from the last 200 response, echoed back verbatim as
-- If-None-Match / If-Modified-Since so unchanged feeds return 304.
ALTER TABLE feeds
    ADD COLUMN etag TEXT,
    ADD COLUMN last_modified TEXT;

-- +goose Down
ALTER TABLE feeds
    DROP COLUMN last_modified,
    DROP COLUMN etag;
