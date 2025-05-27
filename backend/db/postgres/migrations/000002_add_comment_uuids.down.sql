-- Change comment_id to serial from uuid
ALTER TABLE comments DROP COLUMN comment_id;

ALTER TABLE comments ADD COLUMN comment_id SERIAL PRIMARY KEY;

-- Change blacklist_id to serial from uuid
ALTER TABLE blacklist DROP COLUMN blacklist_id;

ALTER TABLE blacklist ADD COLUMN blacklist_id SERIAL PRIMARY KEY;

