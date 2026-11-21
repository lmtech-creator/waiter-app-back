-- Public menu link per restaurant (PDF, Google Drive, Canva, own website).
-- Empty string means no menu configured.

ALTER TABLE restaurants
  ADD COLUMN IF NOT EXISTS menu_url TEXT NOT NULL DEFAULT '';
