package resourceaccess

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/log"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

type ResourceAccessPolicyManager interface {
	AddPolicy(subject, orgID, resource, action string) error
	RemovePolicy(subject, orgID, resource, action string) error
	RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) error
	CheckPermission(subject, orgID, resource, action string) (bool, error)
	AddGroupingPolicy(subject, role, orgID string) error
	RemoveGroupingPolicy(subject, role, orgID string) error
	RefreshPolicies() error
}

type casbinResourceAccessPolicyManager struct {
	enforcer *casbin.SyncedCachedEnforcer
	logger   logger.Logger
	debug    bool
}

type CasbinResourceAccessPolicyManagerConfig struct {
	DBConnectionString     string
	EnableDebugLog         bool
	PolicyAutoLoadInterval time.Duration
	PolicyFilePath         string
}

func NewResourceAccessPolicyManager(cfg CasbinResourceAccessPolicyManagerConfig) (ResourceAccessPolicyManager, error) {
	// Now create the adapter
	// true = use existing database.
	adapter, err := gormadapter.NewAdapter("postgres", cfg.DBConnectionString, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// Create enforcer with our model and the gorm adapter
	enforcer, err := casbin.NewSyncedCachedEnforcer(cfg.PolicyFilePath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Reload the policies from the DB every cfg.PolicyAutoLoadInterval.
	enforcer.StartAutoLoadPolicy(cfg.PolicyAutoLoadInterval)

	// Load policies from database
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load casbin policies: %w", err)
	}
	casbinLogger := &log.DefaultLogger{}
	casbinLogger.EnableLog(cfg.EnableDebugLog)
	enforcer.SetLogger(casbinLogger)

	logger := logger.NewLoggerWithPrefix(context.Background(), "casbin")
	return &casbinResourceAccessPolicyManager{
		enforcer: enforcer,
		logger:   logger,
		debug:    cfg.EnableDebugLog,
	}, nil
}

func (r *casbinResourceAccessPolicyManager) AddPolicy(subject, org, resource, action string) error {
	_, err := r.enforcer.AddPolicy(subject, org, resource, action)
	return err
}

func (r *casbinResourceAccessPolicyManager) RemovePolicy(subject, org, resource, action string) error {
	_, err := r.enforcer.RemovePolicy(subject, org, resource, action)
	return err
}

func (r *casbinResourceAccessPolicyManager) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) error {
	_, err := r.enforcer.RemoveFilteredPolicy(fieldIndex, fieldValues...)
	return err
}

func (r *casbinResourceAccessPolicyManager) CheckPermission(subject, org, resource, action string) (bool, error) {
	if r.debug {
		r.logger.Infof("=== Access Check ===")
		r.logger.Infof("Request: subject=%s, org=%s, resource=%s, action=%s",
			subject, org, resource, action)
	}

	// Check role assignment
	if r.debug {
		roles := r.enforcer.GetRolesForUserInDomain(subject, org)
		r.logger.Infof("User roles: %v", roles)
	}

	ok, err := r.enforcer.Enforce(subject, org, resource, action)
	if r.debug {
		r.logger.Infof("Final decision: %v (err: %v)", ok, err)
	}
	return ok, err
}

func (r *casbinResourceAccessPolicyManager) RefreshPolicies() error {
	if err := r.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load casbin policies: %w", err)
	}
	return nil
}

func (r *casbinResourceAccessPolicyManager) AddGroupingPolicy(subject, role, orgID string) error {
	_, err := r.enforcer.AddGroupingPolicy(subject, role, orgID)
	return err
}

func (r *casbinResourceAccessPolicyManager) RemoveGroupingPolicy(subject, role, orgID string) error {
	_, err := r.enforcer.RemoveGroupingPolicy(subject, role, orgID)
	return err
}
