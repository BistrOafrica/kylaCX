package utils

import (
	"context"
	"fmt"
	"kyla-be/pkg/templates"
	"time"

	"github.com/resend/resend-go/v2"
)

type ResendService struct {
	ApiKey       string
	FromEmail    string
	SupportEmail string
	BaseURL      string
}

func NewResendService(
	apiKey string,
	fromEmail string,
	supportEmail string,
	baseURL string,
) *ResendService {
	return &ResendService{
		ApiKey:       apiKey,
		FromEmail:    fromEmail,
		SupportEmail: supportEmail,
		BaseURL:      baseURL,
	}
}

// SendEmail is a core method to send emails via Resend API
func (rs *ResendService) SendEmail(to []string, subject string, html string) error {
	ctx := context.TODO()
	client := resend.NewClient(rs.ApiKey)

	params := &resend.SendEmailRequest{
		From:    "kyla <noreply@acc.kyla.com>",
		To:      to,
		Subject: subject,
		Html:    html,
	}

	sent, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		// Log error
		fmt.Printf("Error sending email: %v\n", err)
		return fmt.Errorf("error sending email: %v", err)
	}

	// Log success
	fmt.Printf("Email sent with ID: %s\n", sent.Id)
	return nil
}

func (rs *ResendService) SEND_WELCOME_EMAIL(toAddress, username string, firstName string, password string) error {
	// Construct HTML content
	body, err := templates.WELCOME_EMAIL(templates.WelcomeEmailData{
		Name:              username,
		Username:          username,
		Password:          password,
		LoginLinkRedirect: fmt.Sprintf("https://kyla.com/login/?nu=1&e=%s&pw=%s/", toAddress, password),
		ClientEmail:       toAddress,
		SupportEmail:      "support@kyla.com",
		Year:              fmt.Sprintf("%d", time.Now().Year()),
	})

	if err != nil {
		return fmt.Errorf("error generating welcome email template: %v", err)
	}

	dest := []string{toAddress, "info@kyla.com", "joe@kyla.com"}

	return rs.SendEmail(
		dest,
		"Welcome to kyla",
		body,
	)
}

func (rs *ResendService) SEND_APP_TOKEN_AND_SECRET_EMAIL(data templates.AppSecretData) error {
	body, err := templates.APP_SECRET(data)
	if err != nil {
		return fmt.Errorf("error generating app token email template: %v", err)
	}

	dest := []string{data.ClientEmail, "info@kyla.com", "joe@kyla.com"}

	return rs.SendEmail(
		dest,
		"Bonga CX App Token and Secret",
		body,
	)
}

func (rs *ResendService) SEND_NEW_PASSWORD_EMAIL(username, email, newPassword string) error {
	body, err := templates.PASSWORD_RESET_TEMP(templates.ResetPasswordData{
		Name:         username,
		Email:        email,
		Password:     newPassword,
		ClientEmail:  email,
		SupportEmail: rs.SupportEmail,
		Year:         fmt.Sprintf("%d", time.Now().Year()),
	})

	if err != nil {
		return fmt.Errorf("error generating password reset email template: %v", err)
	}

	dest := []string{email, "info@kyla.com", "joe@kyla.com"}

	return rs.SendEmail(
		dest,
		"Bonga CX Password Reset Request",
		body,
	)
}

func (rs *ResendService) SEND_PASSWORD_CHANGE_CONFIRMATION_EMAIL(username, email string) error {
	body, err := templates.CHANGE_PASSWORD_TEMP(templates.ChangePasswordData{
		Name:         username,
		Email:        email,
		ClientEmail:  email,
		SupportEmail: rs.SupportEmail,
		Year:         fmt.Sprintf("%d", time.Now().Year()),
	})

	if err != nil {
		return fmt.Errorf("error generating password change confirmation email template: %v", err)
	}

	dest := []string{email, "info@kyla.com", "joe@kyla.com"}

	return rs.SendEmail(
		dest,
		"Bonga CX Password Change Confirmation",
		body,
	)
}

func (rs *ResendService) SEND_INVITATION_EMAIL(toAddress string, htmlContent string) error {
	dest := []string{toAddress, "info@kyla.com", "joe@kyla.com"}

	return rs.SendEmail(
		dest,
		"You've Been Invited to Join kyla",
		htmlContent,
	)
}
