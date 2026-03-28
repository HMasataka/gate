DELETE FROM refresh_tokens WHERE client_id IS NULL;

ALTER TABLE refresh_tokens
    ALTER COLUMN client_id SET NOT NULL;
