-- +goose Up
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'canceled', 'failed');

-- +goose Down
DROP TYPE IF EXISTS notification_status;
