create table if not exists "users" (
  email varchar not null unique primary key,
  name varchar,
  created_at timestamp default now(),
  updated_at timestamp
);

create index "idx_email_btree" on "users" (email);

create table if not exists "profile_pictures" (
  email varchar references "users" (email) primary key,
  profile_picture varchar not null,
  created_at timestamp default now(),
  updated_at timestamp
);

create table if not exists "messages" (
  id uuid primary key,
  sender varchar not null,
  created_at timestamp default now(),
  updated_at timestamp
);

create unique index "idx_id_sender" on "messages" (id, sender);

create table if not exists "recipients" (
  message_id uuid references "messages" (id),
  recipient varchar not null,
  created_at timestamp default now(),
  updated_at timestamp
);

create unique index "idx_message_recipient" on "recipients" (message_id, recipient);