-- name: GetUrlMapping :one
SELECT * FROM url_mapping
WHERE id = ? LIMIT 1;

-- name: GetUrlMappingByShortcode :one
SELECT * FROM url_mapping
WHERE shortcode = ? LIMIT 1;

-- name: GetUrlMappingByLongurl :one
SELECT * FROM url_mapping
WHERE longurl = ? LIMIT 1;

-- name: ListUrlMapping :many
SELECT * FROM url_mapping
ORDER BY shortcode;

-- name: CreateUrlMapping :one
INSERT INTO url_mapping (
  longurl, shortcode, owner
) VALUES (
  ?, ?, ?
)
RETURNING *;

-- name: UpdateUrlMapping :one
UPDATE url_mapping
set longurl = ?,
shortcode = ?,
owner = ?
WHERE id = ?
RETURNING *;

-- name: DeleteUrlMapping :exec
DELETE FROM url_mapping
WHERE id = ?;
