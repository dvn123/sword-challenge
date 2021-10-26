CREATE TABLE IF NOT EXISTS roles
(
    id   BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users
(
    id       BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    role_id  BIGINT       NOT NULL REFERENCES roles
);

CREATE TABLE IF NOT EXISTS tasks
(
    id             BIGINT           NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id        BIGINT           NOT NULL REFERENCES users,
    summary        VARBINARY(10012) NOT NULL, # 2500*(max char size in UTF-8) + IV
    completed_date TIMESTAMP
);
