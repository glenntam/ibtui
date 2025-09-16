// A simple SMTP client to email messages.
// In this case used to email log messages.
package smtp

import (
	"fmt"
	"net/smtp"
)

// Config settings for simple SMTP client.
type SMTPClient struct {
	Host      string
	Port      int
	Username  string
	Password  string
	Sender    string
	Recipient string
}

// Init a simple SMTP client.
func NewSMTPClient(port int, host, username, password, sender, recipient string) *SMTPClient {
	return &SMTPClient{
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		Sender:    sender,
		Recipient: recipient,
	}
}

// Send a simple email based on previously set config settings.
func (c *SMTPClient) Send(subject, body, recipient string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipient, subject, body)
	return smtp.SendMail(addr, auth, c.Sender, []string{recipient}, []byte(msg))
}
