ALTER TABLE events ADD COLUMN previous_event_code text;
ALTER TABLE events ADD COLUMN previous_code_expiration_time timestamptz;
