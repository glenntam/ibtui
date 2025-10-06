// Package smtp is a simple SMTP client to email log messages.
package smtp

import (
	"fmt"
	"net/smtp"
)

// Client contains config settings for a simple SMTP client.
type Client struct {
	Host      string
	Port      int
	Username  string
	Password  string
	Sender    string
	Recipient string
}

// NewClient initializes a simple SMTP client.
func NewClient(port int, host, username, password, sender, recipient string) *Client {
	return &Client{
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		Sender:    sender,
		Recipient: recipient,
	}
}

// Send a simple email based on previously set config settings.
func (c *Client) Send(subject, body, recipient string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", recipient, subject, body)
	err := smtp.SendMail(addr, auth, c.Sender, []string{recipient}, []byte(msg))
	if err != nil {
		return fmt.Errorf("SMTP Client couldn't send mail. Error: %w", err)
	}
	return nil
}
