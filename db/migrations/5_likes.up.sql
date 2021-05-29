create table if not exists hungries.like
(
    device_id   uuid                               not null,
    place_id    int references hungries.place (id) not null,
    is_liked    boolean                            not null,
    update_date date default now(),
    unique (device_id, place_id)
)
