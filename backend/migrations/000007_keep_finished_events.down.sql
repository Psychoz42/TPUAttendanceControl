DROP INDEX events_owner_created_idx;
DROP INDEX events_previous_code_idx;
ALTER TABLE events DROP COLUMN created_at;
