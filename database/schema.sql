-- Run once in a NEW GameMatch database. These names are isolated from MovieHub.
BEGIN;
CREATE TABLE IF NOT EXISTS gm_members (
 id BIGSERIAL PRIMARY KEY, email TEXT UNIQUE NOT NULL, username TEXT UNIQUE,
 password_hash TEXT NOT NULL, display_name TEXT NOT NULL,
 role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('member','admin')),
 token_version INTEGER NOT NULL DEFAULT 0
);
ALTER TABLE gm_members ADD COLUMN IF NOT EXISTS username TEXT;
UPDATE gm_members SET username=email WHERE username IS NULL;
ALTER TABLE gm_members ALTER COLUMN username SET NOT NULL;
CREATE TABLE IF NOT EXISTS gm_games (
 id BIGSERIAL PRIMARY KEY, rawg_id BIGINT UNIQUE,
 title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', description_th TEXT NOT NULL DEFAULT '', image_url TEXT NOT NULL DEFAULT '',
 genres TEXT[] NOT NULL DEFAULT '{}', tags TEXT[] NOT NULL DEFAULT '{}',
 platforms TEXT[] NOT NULL DEFAULT '{}', is_published BOOLEAN NOT NULL DEFAULT TRUE
);
ALTER TABLE gm_games ADD COLUMN IF NOT EXISTS description_th TEXT NOT NULL DEFAULT '';
-- Safe to rerun when upgrading the original starter database.
ALTER TABLE gm_games ADD COLUMN IF NOT EXISTS steam_id BIGINT UNIQUE;
ALTER TABLE gm_games ADD COLUMN IF NOT EXISTS release_date DATE;
ALTER TABLE gm_games ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE gm_games ADD COLUMN IF NOT EXISTS popularity_order INTEGER;
ALTER TABLE gm_games ADD COLUMN IF NOT EXISTS popularity_source TEXT;
ALTER TABLE gm_games ADD COLUMN IF NOT EXISTS popularity_updated_at TIMESTAMPTZ;
CREATE TABLE IF NOT EXISTS gm_favorites (
 member_id BIGINT REFERENCES gm_members(id) ON DELETE CASCADE,
 game_id BIGINT REFERENCES gm_games(id) ON DELETE CASCADE,
 PRIMARY KEY(member_id,game_id)
);
CREATE TABLE IF NOT EXISTS gm_preferences (
 member_id BIGINT PRIMARY KEY REFERENCES gm_members(id) ON DELETE CASCADE,
 genre TEXT NOT NULL DEFAULT 'rpg', mode TEXT NOT NULL DEFAULT 'solo', style TEXT NOT NULL DEFAULT 'story'
);
CREATE TABLE IF NOT EXISTS gm_reviews (
 id BIGSERIAL PRIMARY KEY,
 member_id BIGINT NOT NULL REFERENCES gm_members(id) ON DELETE CASCADE,
 game_id BIGINT NOT NULL REFERENCES gm_games(id) ON DELETE CASCADE,
 rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
 comment TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE(member_id, game_id)
);
CREATE INDEX IF NOT EXISTS gm_favorites_game_idx ON gm_favorites(game_id);
CREATE INDEX IF NOT EXISTS gm_reviews_game_idx ON gm_reviews(game_id, created_at DESC);
-- Browser never talks to the Supabase Data API. Go uses a trusted server DB role.
ALTER TABLE gm_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE gm_games ENABLE ROW LEVEL SECURITY;
ALTER TABLE gm_favorites ENABLE ROW LEVEL SECURITY;
ALTER TABLE gm_preferences ENABLE ROW LEVEL SECURITY;
ALTER TABLE gm_reviews ENABLE ROW LEVEL SECURITY;
-- No client policies: deny direct anon/authenticated access. Go authorization owns access.
REVOKE ALL ON gm_members,gm_games,gm_favorites,gm_preferences,gm_reviews FROM anon, authenticated;
COMMIT;
-- Make your registered account admin manually (replace email; no default admin password):
-- UPDATE gm_members SET role='admin' WHERE email='your-email@example.com';
