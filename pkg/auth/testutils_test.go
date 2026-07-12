// Package auth_test is an external test package by necessity: these helpers
// import pkg/mocks, which mocks services interfaces and therefore imports
// pkg/services, which imports pkg/auth. In-package test files would close
// that cycle (auth [test] -> mocks -> services -> auth); the external test
// package keeps pkg/auth's own build (and its in-package tests) mocks-free.
package auth_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"go.uber.org/mock/gomock"
)

func newTestPolicyManager() resourceaccess.ResourceAccessPolicyManager {
	pm, err := resourceaccess.NewInMemoryPolicyManager()
	if err != nil {
		panic("failed to create test policy manager: " + err.Error())
	}

	if err := auth.LoadDefaultPolicies(pm.AddPolicy); err != nil {
		panic("failed to load default policies: " + err.Error())
	}

	return pm
}

func ctxWithIdentity(identity *auth.Identity) context.Context {
	return auth.SetIdentityInContext(context.Background(), identity)
}

type testEnv struct {
	policyMgr   resourceaccess.ResourceAccessPolicyManager
	permService auth.PermissionService
	ctrl        *gomock.Controller
}

func newTestEnv(t gomock.TestReporter, projects map[string]*models.Project) *testEnv {
	ctrl := gomock.NewController(t)

	mockProject := mocks.NewMockProjectStore(ctrl)
	for id, project := range projects {
		mockProject.EXPECT().GetByID(gomock.Any(), id).Return(project, nil).AnyTimes()
	}
	mockProject.EXPECT().GetByID(gomock.Any(), gomock.Any()).Return(nil, errors.NotFound("project not found")).AnyTimes()

	mockLog := mocks.NewMockLogger(ctrl)
	mockLog.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLog.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	pm := newTestPolicyManager()
	ps := auth.NewPermissionService(auth.PermissionServiceSpec{
		PolicyManager: pm,
		ProjectStore:  mockProject,
		Logger:        mockLog,
	})
	return &testEnv{policyMgr: pm, permService: ps, ctrl: ctrl}
}
