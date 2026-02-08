-- +goose Up
-- +goose StatementBegin
ALTER TABLE registers ADD COLUMN access TEXT;
ALTER TABLE fields ADD COLUMN access TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE registers DROP COLUMN access;
ALTER TABLE fields DROP COLUMN access;
-- +goose StatementEnd
