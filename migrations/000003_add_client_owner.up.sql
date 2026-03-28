ALTER TABLE oauth_clients
    ADD COLUMN owner_id UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_oauth_clients_owner_id ON oauth_clients (owner_id);
