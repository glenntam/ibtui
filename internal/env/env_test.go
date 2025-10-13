package env

import (
	"os"
	"testing"
)

func TestParseDotEnv_defaults_and_smtp_detection(t *testing.T) {
	// Clear relevant env vars first
	os.Clearenv()
	t.Setenv("IBTUI_HOST", "localhost")
	t.Setenv("IBTUI_PORT", "8080")
	t.Setenv("IBTUI_CLIENT_ID", "42")
	t.Setenv("IBTUI_TIMEZONE", "UTC")
	// Ensure SMTP is not enabled by default
	cfg := ParseDotEnv()
	if cfg.Host != "localhost" {
		t.Fatalf("expected host localhost got %s", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected port 8080 got %d", cfg.Port)
	}
	// Now set SMTP recipient to enable SMTP parsing
	t.Setenv("IBTUI_SMTP_TO", "ops@example.com")
	t.Setenv("IBTUI_SMTP_HOST", "smtp.example.com")
	t.Setenv("IBT_SMTP_PORT", "2525")
	t.Setenv("IBTUI_SMTP_USERNAME", "user")
	t.Setenv("IBTUI_SMTP_PASSWORD", "pass")
	t.Setenv("IBTUI_SMTP_SENDER_EMAIL", "sender@example.com")
	t.Setenv("IBTUI_SMTP_RECIPIENT_EMAIL", "recipient@example.com")

	cfg2 := ParseDotEnv()
	if cfg2.SMTPHost != "smtp.example.com" {
		t.Fatalf("expected smtp host smtp.example.com got %s", cfg2.SMTPHost)
	}
	if cfg2.SMTPPort != 2525 {
		t.Fatalf("expected smtp port 2525 got %d", cfg2.SMTPPort)
	}
}
