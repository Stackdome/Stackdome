package resourceaccess

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/log"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

//go:embed casbin_model.conf
var casbinModelConf string

//go:generate mockgen -destination=../mocks/mock_resource_access_policy_manager.go -package=mocks github.com/Stackdome/stackdome/pkg/resourceaccess ResourceAccessPolicyManager

type ResourceAccessPolicyManager interface {
	AddPolicy(subject, domain, resource, action string) error
	RemovePolicy(subject, domain, resource, action string) error
	RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) error
	CheckPermission(subject, domain, resource, action string) (bool, error)
	AddGroupingPolicy(subject, role, domain string) error
	RemoveGroupingPolicy(subject, role, domain string) error
	HasGroupingPolicy(subject, role, domain string) (bool, error)
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
}

func NewResourceAccessPolicyManager(cfg CasbinResourceAccessPolicyManagerConfig) (ResourceAccessPolicyManager, error) {
	m, err := model.NewModelFromString(casbinModelConf)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded casbin model: %w", err)
	}

	adapter, err := gormadapter.NewAdapter("postgres", cfg.DBConnectionString, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	enforcer, err := casbin.NewSyncedCachedEnforcer(m, adapter)
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

func (r *casbinResourceAccessPolicyManager) AddPolicy(subject, domain, resource, action string) error {
	_, err := r.enforcer.AddPolicy(subject, domain, resource, action)
	return err
}

func (r *casbinResourceAccessPolicyManager) RemovePolicy(subject, domain, resource, action string) error {
	_, err := r.enforcer.RemovePolicy(subject, domain, resource, action)
	return err
}

func (r *casbinResourceAccessPolicyManager) RemoveFilteredPolicy(fieldIndex int, fieldValues ...string) error {
	_, err := r.enforcer.RemoveFilteredPolicy(fieldIndex, fieldValues...)
	return err
}

func (r *casbinResourceAccessPolicyManager) CheckPermission(subject, domain, resource, action string) (bool, error) {
	if r.debug {
		r.logger.Infof("=== Access Check ===")
		r.logger.Infof("Request: subject=%s, domain=%s, resource=%s, action=%s",
			subject, domain, resource, action)
	}

	// Check role assignment
	if r.debug {
		roles := r.enforcer.GetRolesForUserInDomain(subject, domain)
		r.logger.Infof("User roles: %v", roles)
	}

	ok, err := r.enforcer.Enforce(subject, domain, resource, action)
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

func (r *casbinResourceAccessPolicyManager) AddGroupingPolicy(subject, role, domain string) error {
	_, err := r.enforcer.AddGroupingPolicy(subject, role, domain)
	return err
}

func (r *casbinResourceAccessPolicyManager) RemoveGroupingPolicy(subject, role, domain string) error {
	_, err := r.enforcer.RemoveGroupingPolicy(subject, role, domain)
	return err
}

func (r *casbinResourceAccessPolicyManager) HasGroupingPolicy(subject, role, domain string) (bool, error) {
	return r.enforcer.HasGroupingPolicy(subject, role, domain)
}
