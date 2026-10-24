-- +goose Up
Create table refresh_tokens (
    token text primary key,
    created_at timestamp,
    updated_at timestamp,
    user_id uuid not null,
    expires_at timestamp,
    revoked_at timestamp,
    foreign key (user_id) references users(id) on delete cascade
);

-- +goose Down
Drop table refresh_tokens;
