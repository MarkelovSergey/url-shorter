CREATE TABLE IF NOT EXISTS urls (
    uuid VARCHAR(255) PRIMARY KEY,
    short_url VARCHAR(255) NOT NULL UNIQUE,
    original_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_short_url ON urls(short_url);
CREATE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);