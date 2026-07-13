package email

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/logger"
)

type noopEmailService struct {
	logger logger.Logger
}

func NewNoopEmailService(logger logger.Logger) EmailService {
	return &noopEmailService{logger: logger}
}

func (n *noopEmailService) IsConfigured() bool {
	return false
}

func (n *noopEmailService) SendInviteEmail(ctx context.Context, params InviteEmailParams) error {
	n.logger.Infof("email service not configured - invite for %s (org: %s, project: %s) not sent", params.ToEmail, params.OrgName, params.ProjectName)
	return nil
}
