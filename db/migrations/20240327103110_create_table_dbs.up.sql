
CREATE TABLE IF NOT EXISTS dbs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT timezone('utc', now()),
    expires_at TIMESTAMP,
    disabled_at TIMESTAMP,
    is_enabled BOOLEAN DEFAULT TRUE
  );

CREATE UNIQUE INDEX dbs_name_idx ON dbs (name);
