package email

import (
	"context"
	"time"
)

type InviteEmailParams struct {
	ToEmail     string
	OrgName     string
	ProjectName string
	InviterName string
	InviteToken string
	ExpiresAt   time.Time
}

type EmailService interface {
	SendInviteEmail(ctx context.Context, params InviteEmailParams) error
}
