-- Add listing_type field
ALTER TABLE comments ADD listing_type varchar(5);

-- Set all listing types to 'house'
UPDATE comments SET listing_type='house';

-- Set listing_type to not null
ALTER TABLE comments ALTER listing_type SET NOT NULL;