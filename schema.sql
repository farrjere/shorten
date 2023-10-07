CREATE TABLE url_mapping (
  id   INTEGER PRIMARY KEY,
  longurl text NOT NULL,
  shortcode text NOT NULL,
  owner text NOT NULL
);
CREATE UNIQUE INDEX shortcode_idx ON url_mapping (shortcode);
CREATE UNIQUE INDEX longurl_idx ON url_mapping (longurl);
