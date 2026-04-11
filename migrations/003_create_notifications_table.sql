-- +goose Up
CREATE TABLE notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message    TEXT                NOT NULL,
    send_at    TIMESTAMP           NOT NULL,
    status     notification_status NOT NULL DEFAULT 'pending',
    retries    INT                 NOT NULL DEFAULT 0 CHECK (retries >= 0),
    recipient  TEXT                NOT NULL,
    channel    TEXT                NOT NULL CHECK (channel IN ('email', 'telegram')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_notifications_status_send_at 
    ON notifications(status, send_at) 
    WHERE status = 'pending';

-- +goose Down
DROP TABLE notifications;
