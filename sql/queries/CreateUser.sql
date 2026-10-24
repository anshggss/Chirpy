-- name: CreateUser :one
insert into 
users (
    id,
    created_at,
    updated_at,
    email,
    hashed_password
) values (
$1,
NOW(),
NOW(),
$2,
$3
)
returning *;
