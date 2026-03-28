-- Allow receiver to delete messages (hide from their own view)
ALTER TABLE messages ADD COLUMN IF NOT EXISTS deleted_by_receiver BOOLEAN NOT NULL DEFAULT false;
