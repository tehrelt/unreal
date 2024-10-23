create table if not exists "known_addresses" (
  host varchar not null primary key,
  picture varchar not null,
  created_at TIMESTAMP default now(),
  updated_at timestamp
)