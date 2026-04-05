-- name: ListFeeds :many
SELECT feeds.name, feeds.url, users.name, feeds.id
FROM feeds
JOIN users ON feeds.user_id = users.id;