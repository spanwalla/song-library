CREATE INDEX IF NOT EXISTS idx_songs_song_name_hash ON songs USING HASH (song_name);

CREATE INDEX IF NOT EXISTS idx_songs_group_name_hash ON songs USING HASH (group_name);