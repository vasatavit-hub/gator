-- name: GetFeedFollowsForUser :many
SELECT feeds.name AS feed_name, users.name AS user_name, feeds.id AS feed_id
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;