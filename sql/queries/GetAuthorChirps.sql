-- name: GetAuthorChirps :many
Select * from chirps
where user_id = $1 order by created_at;
