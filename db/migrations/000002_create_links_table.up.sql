CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    original_url TEXT NOT NULL,
    slug VARCHAR(50) NOT NULL,
    click_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_links_user_id ON links(user_id);
CREATE UNIQUE INDEX idx_links_slug_active ON links(slug) WHERE deleted_at IS NULL;