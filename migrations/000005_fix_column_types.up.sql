ALTER TABLE events ALTER COLUMN expiration_time TYPE timestamptz USING (CURRENT_DATE + expiration_time);

ALTER TABLE events ALTER COLUMN noted_students TYPE jsonb USING COALESCE(noted_students::jsonb, '[]'::jsonb);
ALTER TABLE events ALTER COLUMN noted_students SET DEFAULT '[]'::jsonb;
UPDATE events SET noted_students = '[]'::jsonb WHERE noted_students IS NULL;
ALTER TABLE events ALTER COLUMN noted_students SET NOT NULL;

UPDATE events SET update_need = false WHERE update_need IS NULL;
ALTER TABLE events ALTER COLUMN update_need SET DEFAULT false;
ALTER TABLE events ALTER COLUMN update_need SET NOT NULL;
