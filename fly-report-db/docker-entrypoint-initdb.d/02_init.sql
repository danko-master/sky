---- Создаём дополнения
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
CREATE EXTENSION IF NOT EXISTS postgis;


---- Создаём таблицы
CREATE TABLE regions (
    id SERIAL PRIMARY KEY,
    rus_code integer,
    rus_name VARCHAR(255),
    geometry_geojson JSONB,
    geometry_postgis geometry(Geometry,4326)
);

CREATE TABLE orvd (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255)
);


CREATE TABLE flights (
    id SERIAL,
    region_id integer,
    orvd_id integer,
    coords_postgis geometry(Geometry,4326) not null,
    event_date DATE not null,
    xlsxfile_uuid UUID,
    row_num integer,
    details JSONB not null,
    details_SHA256 VARCHAR(64) not null,
    "createdAt"    TIMESTAMP DEFAULT NOW(),
    "updatedAt"    TIMESTAMP DEFAULT NOW()
);

CREATE UNIQUE INDEX unique_event_date_id ON flights (event_date, id);
ALTER TABLE flights ADD CONSTRAINT unique_flights_event_date_id UNIQUE (event_date, id);
CREATE UNIQUE INDEX unique_event_date_details ON flights (event_date, details_SHA256);
ALTER TABLE flights ADD CONSTRAINT unique_flights_event_date_details UNIQUE (event_date, details_SHA256);
CREATE INDEX event_date_region_id ON flights (event_date, region_id);
CREATE INDEX event_date_xlsxfile_uuid ON flights (event_date, xlsxfile_uuid);



