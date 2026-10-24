-- +goose Up
create table 
users (
    id UUID primary key,
    created_at timestamp,
    updated_at timestamp,
    email text not null unique,
    hashed_password text not null,
    is_chirpy_red boolean default false
);

-- +goose Down
drop table users;
