-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetFeedID :one

SELECT id
FROM feeds
WHERE url = $1 LIMIT 1;

-- name: GetFeeds :many

SELECT feeds.name, feeds.url, 
users.name as creator
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
    ORDER BY last_fetched_at ASC NULLS FIRST
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;
