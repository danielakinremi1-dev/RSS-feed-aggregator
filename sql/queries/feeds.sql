-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFeedID :one

SELECT id
FROM feeds
WHERE url = $1 LIMIT 1;

-- name: GetFeeds :many

SELECT feeds.name, feeds.url, users.name AS creator,
    feeds.consecutive_failures, feeds.next_fetch_at
FROM feeds
INNER JOIN users on users.id = feeds.user_id;

-- name: ClaimFeedsToFetch :many
-- Atomically claims up to $1 of the stalest feeds by stamping last_fetched_at.
-- FOR UPDATE SKIP LOCKED makes concurrent `agg` processes claim disjoint sets:
-- rows another claim has locked are skipped instead of waited on. This must be
-- a single statement; a bare SELECT ... FOR UPDATE releases its locks as soon
-- as it returns (autocommit), before any fetching happens.
UPDATE feeds
SET last_fetched_at = now(), updated_at = now()
WHERE id IN (
    SELECT id FROM feeds
    WHERE next_fetch_at IS NULL OR next_fetch_at <= now()
    ORDER BY last_fetched_at ASC NULLS FIRST
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: RecordFeedSuccess :exec
UPDATE feeds
SET consecutive_failures = 0, last_error = NULL, next_fetch_at = NULL, updated_at = now()
WHERE id = $1;

-- name: RecordFeedFailure :exec
-- The delay is computed in Go (exponential backoff); the timestamp uses the DB clock.
UPDATE feeds
SET consecutive_failures = consecutive_failures + 1,
    last_error = sqlc.arg(last_error),
    next_fetch_at = now() + make_interval(secs => sqlc.arg(retry_after_seconds)::float8),
    updated_at = now()
WHERE id = sqlc.arg(id);
