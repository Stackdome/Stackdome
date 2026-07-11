package email

import (
	"context"
	"fmt"
	"html"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host        string
	Port        string
	Username    string
	Password    string
	FromAddress string
	AppBaseURL  string
}

type smtpEmailService struct {
	config SMTPConfig
}

func NewSMTPEmailService(config SMTPConfig) EmailService {
	return &smtpEmailService{config: config}
}

func (s *smtpEmailService) IsConfigured() bool {
	return true
}

func (s *smtpEmailService) SendInviteEmail(ctx context.Context, params InviteEmailParams) error {
	inviteURL := fmt.Sprintf("%s/invite/%s", strings.TrimRight(s.config.AppBaseURL, "/"), params.InviteToken)

	orgName := html.EscapeString(params.OrgName)
	inviterName := html.EscapeString(params.InviterName)
	projectName := html.EscapeString(params.ProjectName)

	subjectOrgName := strings.NewReplacer("\r", "", "\n", "").Replace(params.OrgName)
	subject := fmt.Sprintf("You've been invited to join %s", subjectOrgName)
	body := fmt.Sprintf(`<html><body>
<h2>You've been invited to join %s</h2>
<p>%s has invited you to join the <strong>%s</strong> project in the <strong>%s</strong> organization.</p>
<p><a href="%s">Accept Invitation</a></p>
<p>This invitation expires on %s.</p>
<p>If you didn't expect this invitation, you can safely ignore this email.</p>
</body></html>`,
		orgName,
		inviterName,
		projectName,
		orgName,
		html.EscapeString(inviteURL),
		params.ExpiresAt.Format("January 2, 2006"),
	)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.config.FromAddress, params.ToEmail, subject, body)

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	return smtp.SendMail(addr, auth, s.config.FromAddress, []string{params.ToEmail}, []byte(msg))
}
