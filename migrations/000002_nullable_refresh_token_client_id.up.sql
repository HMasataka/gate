ALTER TABLE refresh_tokens
    ALTER COLUMN client_id DROP NOT NULL,
    ALTER COLUMN client_id SET DEFAULT NULL;

ALTER TABLE refresh_tokens
    DROP CONSTRAINT IF EXISTS refresh_tokens_client_id_fkey;

ALTER TABLE refresh_tokens
    ADD CONSTRAINT refresh_tokens_client_id_fkey
    FOREIGN KEY (client_id) REFERENCES oauth_clients(id) ON DELETE CASCADE;
