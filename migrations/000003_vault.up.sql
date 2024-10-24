create table if not exists "vault" (
  "id" uuid primary key not null,
  "key" varchar not null,
  "hashsum" varchar not null,
  "created_at" timestamp default now(),
  "updated_at" timestamp
)