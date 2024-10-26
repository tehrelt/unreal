create table if not exists vault_files (
  id uuid primary key,
  message_id uuid not null references vault(id) on delete cascade,
  file_name text not null,
  content_type varchar not null,
  key varchar not null,
  hashsum varchar,
  created_at timestamp default now(),
  updated_at timestamp
);