package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host          string
	Port          int
	ClientID      int64
	Timezone      string
	LogFile       string
	SMTPHost      string
	SMTPPort      int
	SMTPUsername  string
	SMTPPassword  []byte
	SMTPSender    string
	SMTPRecipient string
}

// Get environment variables from .env file.
//
// A file called ".env" must be located where the binary is run.
// If .env exists but required environment variables are not found, then reasonable default values are used.
func ParseDotEnv() *Config {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(".env file not found! Please create one. \n\n A sample template file called .env-example is included for your reference")
		os.Exit(1)
	}

	host := os.Getenv("IBTUI_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("IBTUI_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 4001
	}

	clientIDStr := os.Getenv("IBTUI_CLIENT_ID")
	clientIDInt, err := strconv.Atoi(clientIDStr)
	if err != nil {
		clientIDInt = 0
	}
	clientID := int64(clientIDInt)

	timezone := os.Getenv("IBTUI_TIMEZONE")
	if timezone == "" {
		timezone = "UTC"
	}

	logFile := os.Getenv("IBTUI_LOG_FILE")
	if logFile == "" {
		logFile = "logfile.json"
	}

	cfg := &Config{
		Host:        host,
		Port:        port,
		ClientID:    clientID,
		Timezone:    timezone,
		LogFile:     logFile,
	}

	smtpTo := os.Getenv("IBTUI_SMTP_TO")
	// Assume user wants email capability if IBT_SMTP_TO is filled in
	if smtpTo != "" && smtpTo != "admin@example.com" {
		cfg.SMTPHost = os.Getenv("IBTUI_SMTP_HOST")
		if cfg.SMTPPort, err = strconv.Atoi(os.Getenv("IBT_SMTP_PORT")); err != nil {
			cfg.SMTPPort = 456
		}
		cfg.SMTPUsername = os.Getenv("IBTUI_SMTP_USERNAME")
		cfg.SMTPPassword = []byte(os.Getenv("IBTUI_SMTP_PASSWORD"))
		cfg.SMTPSender = os.Getenv("IBTUI_SMTP_SENDER_EMAIL")
		cfg.SMTPRecipient = os.Getenv("IBTUI_SMTP_RECIPIENT_EMAIL")
	}
	return cfg
}
