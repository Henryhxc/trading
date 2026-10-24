DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
DELETE FROM schema_migrations WHERE version = '001_users_sessions_accounts.sql';
