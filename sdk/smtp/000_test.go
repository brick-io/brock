// Unit Test Code
package smtp_test

import (
	"testing"

	smtp "github.com/brick-io/brock/sdk/smtp"
)

func TestSendEmail(t *testing.T) {
	// Create a mock SMTPConfiguration
	mockSMTPConfig := smtp.SMTPConfiguration{
		Host:         "",
		Port:         2525,
		AuthUsername: "",
		AuthPassword: "",
		Sender:       "",
	}

	// Create a mock recipient, cc, subject, body and attachmentPath strings
	recipient := []string{"recipient@example.com"}
	cc := []string{"cc@example.com"}
	subject := "Test Email"
	body := "This is a test email"
	attachmentPath := ""

	// Call SendEmail with the mock data and useHTML set to true
	err := mockSMTPConfig.SendEmail(recipient, cc, subject, body, attachmentPath, true)
	// Assert that no error was returned
	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}

	// Call SendEmail with the mock data and useHTML set to false
	err = mockSMTPConfig.SendEmail(recipient, cc, subject, body, attachmentPath, false)

	// Assert that no error was returned
	if err != nil {
		t.Errorf("Expected no error but got %v", err)
	}

	// Call SendEmail with an invalid attachment path and useHTML set to true
	err = mockSMTPConfig.SendEmail(recipient, cc, subject, body, "/invalid/path/to/attachment", true)

	// Assert that an error was returned
	if err == nil {
		t.Errorf("Expected an error but got none")
	}
}
