create table if not exists "users" (
  email varchar not null unique primary key,
  name varchar,
  created_at timestamp default now(),
  updated_at timestamp
);

create index "idx_email_btree" on "users" (email);

create table if not exists "profile_pictures" (
  email varchar primary key references "users" (email) on delete cascade,
  profile_picture varchar not null,
  created_at timestamp default now(),
  updated_at timestamp
);