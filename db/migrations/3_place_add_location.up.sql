create extension if not exists postgis;

alter table hungries.place
    add column if not exists location geography(Point);