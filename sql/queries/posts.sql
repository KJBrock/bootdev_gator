-- name: CreatePost :one
INSERT INTO posts(
    id, 
    created_at, 
    updated_at, 
    published_at,
    title, 
    url, 
    description,
    feed_id)
VALUES($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.* 
FROM posts 
LEFT JOIN feed_follows ON feed_follows.user_id = $1
WHERE posts.feed_id = feed_follows.feed_id
ORDER BY published_at DESC
LIMIT $2; 
