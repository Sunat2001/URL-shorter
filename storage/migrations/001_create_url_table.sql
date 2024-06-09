CREATE TABLE IF NOT EXISTS url
(
    id    INTEGER PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    url   TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    CONSTRAINT foreign_url_user_id FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS idx_alias ON url (alias);