-- Эта миграция создаёт все необходимые таблицы и индексы. 
-- Обратите внимание, что параметры email, name и secret должны быть уникальными 
-- и мы добавили для них соответствующий constraint.
CREATE TABLE IF NOT EXISTS users
(
id          INTEGER PRIMARY KEY,
email       TEXT    NOT NULL UNIQUE,    --должно быть уникальным
pass_hash   BLOB    NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_email ON users (email);

CREATE TABLE IF NOT EXISTS apps
(
    id     INTEGER PRIMARY KEY,
    name   TEXT NOT NULL UNIQUE,        --должно быть уникальным
    secret TEXT NOT NULL UNIQUE         --должно быть уникальным
);