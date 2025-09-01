-- Add ip_nonce [varchar(24)] field from comments table to allow for encryption of ips
-- Is nullable because it wasn't included in previous versions, should be changed later
ALTER TABLE comments ADD ip_nonce varchar(24); 

-- Change user_ip to varchar(130) from varchar(45), accounting for encryption bloat
ALTER TABLE comments ALTER COLUMN user_ip TYPE varchar(130);