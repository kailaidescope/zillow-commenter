-- name: GetCommentsByListingID :many
SELECT comment_id, listing_id, user_ip, user_id, username, comment_text, EXTRACT(EPOCH FROM date_created) FROM comments
WHERE listing_id = $1
ORDER BY date_created DESC;

-- name: PostComment :one
INSERT INTO comments (comment_id, listing_id, user_ip, user_id, username, comment_text)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING comment_id, listing_id, user_ip, user_id, username, comment_text, EXTRACT(EPOCH FROM date_created);