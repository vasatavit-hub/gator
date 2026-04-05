-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = NOW()
WHERE id = $1;