// Package sender provides infrastructure logic for sending notifications via various channels.
package sender

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/akhmed9505/delayed-notifier/internal/config"
)

// SMTPMailer implements the mailer interface using the SMTP protocol.
type SMTPMailer struct {
	cfg *config.SMTP
}

// NewSMTPMailer creates a new instance of SMTPMailer with the provided SMTP configuration.
func NewSMTPMailer(cfg *config.SMTP) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

// Send transmits a message to the specified recipient using the SMTP protocol.
func (s *SMTPMailer) Send(ctx context.Context, message, recipient string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() {
		_ = c.Quit()
	}()

	if s.cfg.UseTLS {
		tlsCfg := &tls.Config{
			ServerName: s.cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	if err := c.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}

	if err := c.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	if err := c.Rcpt(recipient); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	wc, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	msg := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: Notification\r\n\r\n%s",
		recipient, message,
	))

	if _, err := wc.Write(msg); err != nil {
		_ = wc.Close()
		return fmt.Errorf("smtp write: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}

	return nil
}
