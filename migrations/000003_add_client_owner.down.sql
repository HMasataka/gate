DROP INDEX IF EXISTS idx_oauth_clients_owner_id;

ALTER TABLE oauth_clients
    DROP COLUMN IF EXISTS owner_id;
