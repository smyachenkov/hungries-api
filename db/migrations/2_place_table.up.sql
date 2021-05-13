create table if not exists hungries.place
(
    id              serial primary key,
    google_place_id text unique not null,
    name            text        not null,
    url             text        not null
);
