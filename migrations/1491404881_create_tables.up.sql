CREATE TABLE gists(
  id varchar(256) PRIMARY KEY,
  url varchar(2083),
  description text,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  last_loaded_at timestamp NOT NULL
);

CREATE TABLE files(
  filename text NOT NULL,
  type varchar(256),
  language varchar(256),
  gist_id varchar(256) NOT NULL references gists(id)
);

CREATE TABLE tags(
  id serial PRIMARY KEY,
  name varchar(256)
);

CREATE TABLE gist_tag(
  gist_id varchar(256) NOT NULL references gists(id),
  tag_id integer NOT NULL references tags(id)
);
