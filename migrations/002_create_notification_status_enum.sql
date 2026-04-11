-- +goose Up
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'cancelled', 'failed');

-- +goose Down
DROP TYPE IF EXISTS notification_status;
