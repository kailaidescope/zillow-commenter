-- Remove ip_nonce [varchar(24)] field from comments table
ALTER TABLE comments DROP COLUMN ip_nonce;

-- Change user_ip to varchar(45) from varchar(130), assuming encryption is no longer active
ALTER TABLE comments ALTER COLUMN user_ip TYPE varchar(45);