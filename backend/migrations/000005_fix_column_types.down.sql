ALTER TABLE events ALTER COLUMN update_need DROP NOT NULL;
ALTER TABLE events ALTER COLUMN update_need DROP DEFAULT;

ALTER TABLE events ALTER COLUMN noted_students DROP NOT NULL;
ALTER TABLE events ALTER COLUMN noted_students DROP DEFAULT;
ALTER TABLE events ALTER COLUMN noted_students TYPE json USING noted_students::json;

ALTER TABLE events ALTER COLUMN expiration_time TYPE time USING expiration_time::time;
