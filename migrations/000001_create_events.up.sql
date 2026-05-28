create table events (
    event_id serial primary key,
    event_name text,
    event_owner text,
    event_info text,
    event_code text,
    expiration_time time,
    update_need boolean,
    noted_students json
)