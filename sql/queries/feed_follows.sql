-- name: CreateFeedFollow :one
WITH feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
        VALUES (
            $1,
            $2,
            $3,
            $4,
            $5) RETURNING *
    ) SELECT feed_follow.*, users.name AS UserName, feeds.name AS FeedName 
        FROM feed_follow
            JOIN users ON user_id = users.id
            JOIN feeds ON feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
WITH their_feeds AS (
    SELECT * FROM feed_follows
    WHERE feed_follows.user_id = $1
) SELECT their_feeds.*, users.name AS UserName, feeds.name AS FeedName 
        FROM their_feeds
            JOIN users ON user_id = users.id
            JOIN feeds ON their_feeds.feed_id = feeds.id;
