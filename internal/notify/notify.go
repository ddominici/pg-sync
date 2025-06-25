package notify

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"

	"github.com/ddominici/pg-sync/internal/config"
	"github.com/jordan-wright/email"
)

func SendErrorEmail(cfg config.EmailConfig, subject, body string) error {
	if !cfg.Enabled {
		return nil
	}

	e := email.NewEmail()
	e.From = cfg.From
	e.To = cfg.To
	e.Subject = subject
	e.Text = []byte(body)

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	// Dial SMTP server manually
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// STARTTLS if available
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false, // true only for test/self-signed
			ServerName:         cfg.SMTPHost,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
	}

	// Set sender and recipients
	if err := client.Mail(cfg.From); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	for _, to := range cfg.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", to, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command failed: %w", err)
	}

	msg, err := e.Bytes()
	if err != nil {
		return fmt.Errorf("failed to generate message: %w", err)
	}

	if _, err := writer.Write(msg); err != nil {
		return fmt.Errorf("writing email body failed: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("closing DATA stream failed: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("SMTP QUIT failed: %w", err)
	}

	return nil
}
