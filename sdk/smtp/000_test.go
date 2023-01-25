// Unit Test Code
package smtp_test

import (
	"os"
	"strconv"
	"testing"

	smtp "github.com/brick-io/brock/sdk/smtp"
)

func TestSendEmail(t *testing.T) {
	// Create a mock SMTPConfiguration
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	mockSMTPConfig := smtp.SMTPConfiguration{
		Host:         os.Getenv("SMTP_HOST"),
		Port:         port,
		AuthUsername: os.Getenv("SMTP_USERNAME"),
		AuthPassword: os.Getenv("SMTP_PASSWORD"),
		Sender:       "no-reply@onebrick.io",
	}

	// Create a mock recipient, cc, subject, body and attachmentPath strings
	recipient := []string{"recipient@example.com"}
	cc := []string{"cc@example.com"}
	subject := "Test Email"
	body := "This is a test email"
	bodyHTML := "This is a test email"
	attachmentPath := "<html><body>This is a test email<body></html>"

	// Call SendEmail with the mock data and useHTML set to true
	err := mockSMTPConfig.SendEmail(recipient, cc, subject, bodyHTML, attachmentPath)
	// Assert that no error was returned
	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}

	// Call SendEmail with the mock data and useHTML set to false
	err = mockSMTPConfig.SendEmail(recipient, cc, subject, body, attachmentPath)

	// Assert that no error was returned
	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}

	// Call SendEmail with an invalid attachment path and useHTML set to true
	err = mockSMTPConfig.SendEmail(recipient, cc, subject, bodyHTML, "/invalid/path/to/attachment")

	// Assert that an error was returned
	if err == nil {
		t.Errorf("Expected an error but got none")
	}
}
