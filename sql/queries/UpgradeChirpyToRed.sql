-- name: UpgradeChirpyToRed :exec
update users
set is_chirpy_red = true 
where id = $1;
