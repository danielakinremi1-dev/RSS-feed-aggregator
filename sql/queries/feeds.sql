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

-- name: MarkFeedFetched :exec

UPDATE feeds
SET updated_at = NOW(), last_fetched_at = NOW()
WHERE id = $1;

-- name: GetNextFeedToFetch :one

SELECT *
FROM feeds 
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;