-- +goose Up
-- +goose StatementBegin
CREATE TABLE enumerations (id SERIAL PRIMARY KEY, field_id integer REFERENCES fields NOT NULL, name text NOT NULL, value text NOT NULL, description text, UNIQUE (field_id, name, value));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE enumerations;
-- +goose StatementEnd
