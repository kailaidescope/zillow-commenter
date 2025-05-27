-- Drop serial comment_id column
ALTER TABLE comments
DROP COLUMN comment_id;

-- Add comment_id as uuid, primary key, with default value gen_random_uuid()
ALTER TABLE comments
ADD COLUMN comment_id uuid PRIMARY KEY;

-- Drop serial blacklist_id column
ALTER TABLE blacklist
DROP COLUMN blacklist_id;

-- Add blacklist_id as uuid, primary key, with default value gen_random_uuid()
ALTER TABLE blacklist
ADD COLUMN blacklist_id uuid PRIMARY KEY;
