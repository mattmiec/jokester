create table if not exists users (
    user_id uuid not null primary key default gen_random_uuid(),
    openid_sub text not null unique,
    username text not null
);

create table if not exists jokes (
    joke_id uuid not null primary key default gen_random_uuid(),
    author_id uuid not null references users(user_id) on delete cascade,
    created timestamp with time zone not null,
    joke text not null
);

create table if not exists likes (
    user_id uuid not null references users(user_id) on delete cascade,
    joke_id uuid not null references jokes(joke_id) on delete cascade,
    primary key(user_id, joke_id)
);