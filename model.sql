drop table if exists item;
create table item (
  item_id serial primary key,
  netflix_id int unique,
  imdb_id text,
  title text,
  summary text,
  item_type text,
  year int,
  api_date date,
  duration text,
  image_url text,
  image bytea
);

drop table if exists user_account;
create table user_account (
  user_account_id serial primary key,
  email text unique,
  is_active bool default true
);
