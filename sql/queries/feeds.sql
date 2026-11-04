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
-- Atomically claims up to $1 of the stalest due feeds.
--
-- FOR UPDATE SKIP LOCKED makes concurrent `agg` processes skip rows another
-- claim is holding instead of waiting on them. It must be one statement: a
-- bare SELECT ... FOR UPDATE releases its locks as soon as it returns.
--
-- Setting next_fetch_at is a lease, and it is what makes claims disjoint:
-- if another claim commits a row after this statement's snapshot, Postgres
-- re-checks only the WHERE clause against the committed row, and the future
-- next_fetch_at now fails it. It also lets a feed be retried if the claiming
-- process dies mid-fetch. RecordFeedSuccess/RecordFeedFailure overwrite it.
UPDATE feeds
SET last_fetched_at = now(),
    next_fetch_at = now() + interval '5 minutes',
    updated_at = now()
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
SET consecutive_failures = 0, last_error = NULL, next_fetch_at = NULL,
    etag = sqlc.narg(etag), last_modified = sqlc.narg(last_modified),
    updated_at = now()
WHERE id = sqlc.arg(id);

-- name: RecordFeedFailure :exec
-- The delay is computed in Go (exponential backoff); the timestamp uses the DB clock.
UPDATE feeds
SET consecutive_failures = consecutive_failures + 1,
    last_error = sqlc.arg(last_error),
    next_fetch_at = now() + make_interval(secs => sqlc.arg(retry_after_seconds)::float8),
    updated_at = now()
WHERE id = sqlc.arg(id);
