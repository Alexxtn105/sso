-- 2_add_is_admin_column_to_users_tbl.up.sql
ALTER TABLE users
    ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE;
