-- name: UpdateUser :one
update users
set email = $1, hashed_password =$2 , updated_at = $3 where id = $4 returning *;
