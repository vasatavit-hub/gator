-- name: GetNextFeedToFetch :many
SELECT *
FROM feeds
ORDER BY last_fetched_at DESC NULLS FIRST;