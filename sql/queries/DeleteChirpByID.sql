-- name: DeleteChirpByID :exec
delete from chirps where id = $1;
