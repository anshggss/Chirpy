-- name: RevokeToken :exec
update refresh_tokens 
set updated_at = $1, revoked_at = $2
where token = $3;
