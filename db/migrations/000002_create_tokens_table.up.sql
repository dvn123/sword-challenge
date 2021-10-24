CREATE TABLE IF NOT EXISTS tokens
(
    uuid         VARCHAR(128) UNIQUE NOT NULL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users,
    created_date TIMESTAMP           NOT NULL
);
