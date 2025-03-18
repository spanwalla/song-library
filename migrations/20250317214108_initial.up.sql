CREATE TABLE songs(
    id SERIAL PRIMARY KEY,
    song_name VARCHAR(128) NOT NULL,
    group_name VARCHAR(128) NOT NULL,
    link VARCHAR(128) NOT NULL,
    release_date DATE NOT NULL
);

CREATE TABLE couplets(
    song_id INTEGER NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    sequence_number INTEGER NOT NULL,
    couplet_text BPCHAR NOT NULL,
    PRIMARY KEY (song_id, sequence_number)
);