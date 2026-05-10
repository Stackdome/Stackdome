package resourceaccess

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

func NewInMemoryPolicyManager() (ResourceAccessPolicyManager, error) {
	m, err := model.NewModelFromString(casbinModelConf)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewSyncedCachedEnforcer(m)
	if err != nil {
		return nil, err
	}

	return &casbinResourceAccessPolicyManager{
		enforcer: enforcer,
		logger:   logger.NewLoggerWithPrefix(context.Background(), "casbin-test"),
		debug:    false,
	}, nil
}
