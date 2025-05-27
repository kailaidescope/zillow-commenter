CREATE TABLE IF NOT EXISTS comments (
    comment_id UUID PRIMARY KEY,
    listing_id varchar(200) NOT NULL,
    user_ip varchar(45) NOT NULL,
    user_id varchar(50) NOT NULL,
    username varchar(50) NOT NULL,
    comment_text varchar(300) NOT NULL,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blacklist (
    blacklist_id UUID PRIMARY KEY,
    cause varchar(100) NOT NULL,
    user_ip varchar(45) NOT NULL,
    user_id varchar(50),
    username varchar(50),
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);