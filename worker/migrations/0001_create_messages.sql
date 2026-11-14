CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	encrypted TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	expires_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS messages_expires_at ON messages (expires_at);
