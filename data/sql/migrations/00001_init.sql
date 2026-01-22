-- +goose Up
-- +goose StatementBegin
CREATE TABLE mpus (id SERIAL PRIMARY KEY, name text NOT NULL UNIQUE, description text);
CREATE TABLE peripherals (id SERIAL PRIMARY KEY, mpu_id integer REFERENCES mpus NOT NULL, derived_from_id integer, name text NOT NULL, base_address text NOT NULL, description text, UNIQUE (mpu_id, name));
CREATE TABLE registers (id SERIAL PRIMARY KEY, peripheral_id integer REFERENCES peripherals NOT NULL, name text NOT NULL, address_offset text NOT NULL, reset_value text, description text);
CREATE TABLE fields (id SERIAL PRIMARY KEY, register_id integer REFERENCES registers NOT NULL, name text NOT NULL, num_bits integer NOT NULL, bit_offset integer NOT NULL, description text);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE fields;
DROP TABLE registers;
DROP TABLE peripherals;
DROP TABLE mpus;
-- +goose StatementEnd
