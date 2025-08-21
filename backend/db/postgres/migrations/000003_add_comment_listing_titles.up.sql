-- Add listing_title field to comments table
-- listing_title is nullable because there exist rows with no listing title, should be changed later
ALTER TABLE comments ADD COLUMN listing_title varchar(200);