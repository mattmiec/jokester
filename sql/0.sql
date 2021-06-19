create table if not exists users (
    user_id uuid not null primary key default gen_random_uuid(),
    username text not null,
    email text not null
);

create table if not exists jokes (
    joke_id uuid not null primary key default gen_random_uuid(),
    author_id uuid not null references users(user_id),
    created timestamp with time zone not null,
    joke text not null
);

create table if not exists likes (
    user_id uuid not null references users(user_id),
    joke_id uuid not null references jokes(joke_id)
);