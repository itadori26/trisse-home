-- Sessions get a public identifier (shown in the session overview instead of
-- the token hash) and the client they were created from.
ALTER TABLE auth_sessions ADD COLUMN public_id TEXT NOT NULL DEFAULT '';
ALTER TABLE auth_sessions ADD COLUMN user_agent TEXT NOT NULL DEFAULT '';
ALTER TABLE auth_sessions ADD COLUMN client_ip TEXT NOT NULL DEFAULT '';
UPDATE auth_sessions SET public_id = lower(hex(randomblob(16))) WHERE public_id = '';
CREATE UNIQUE INDEX auth_sessions_public_id ON auth_sessions (public_id);
