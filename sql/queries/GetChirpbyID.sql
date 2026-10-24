-- name: GetChirpbyID :one
select * from chirps where id = $1;
