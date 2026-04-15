-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetFeedByUrl :one
SELECT * FROM feeds 
WHERE feeds.url = $1;


-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_updated_at=CURRENT_TIMESTAMP
WHERE id=$1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_updated_at ASC NULLS FIRST
LIMIT 1;

