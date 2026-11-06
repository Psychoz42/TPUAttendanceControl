ALTER TABLE events ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();

-- закрытые события теперь остаются в таблице, поэтому поиск не должен просматривать её целиком
CREATE INDEX events_previous_code_idx ON events (previous_event_code) WHERE update_need;
CREATE INDEX events_owner_created_idx ON events (event_owner, created_at DESC);
