-- name: CreateChirp :one
insert into chirps
(
    id,
    created_at,
    updated_at,
    body,
    user_id
) values (
$1,
NOW(),
NOW(),
$2,
$3
) returning *;
