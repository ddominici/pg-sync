package notify

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/ddominici/pg-sync/internal/config"
)

func SendErrorEmail(emailCfg config.EmailConfig, subject, body string) error {
	if !emailCfg.Enabled {
		return nil
	}

	auth := smtp.PlainAuth("", emailCfg.Username, emailCfg.Password, emailCfg.SMTPHost)

	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(emailCfg.To, ", "), subject, body)

	addr := fmt.Sprintf("%s:%d", emailCfg.SMTPHost, emailCfg.SMTPPort)
	return smtp.SendMail(addr, auth, emailCfg.From, emailCfg.To, []byte(msg))
}
