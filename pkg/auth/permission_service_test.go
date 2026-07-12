package auth_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

func TestPermissionService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PermissionService Suite")
}

const (
	projectABC = "project-abc"
	projectXYZ = "project-xyz"
	orgABC     = "org-abc"
	orgXYZ     = "org-xyz"
	userDev    = "user-dev"
	userView   = "user-view"
	userMem    = "user-mem"
	userAdmin  = "user-admin"
)

type permCheck struct {
	domain      string
	resource    string
	resourceID  string
	action      string
	shouldAllow bool
}

type tokenPermCheck struct {
	scopes      []string
	resourceIDs []string
	domain      string
	resource    string
	resourceID  string
	action      string
	shouldAllow bool
}

func defaultProjects() map[string]*models.Project {
	return map[string]*models.Project{
		projectABC: {ID: projectABC, OrganisationID: orgABC},
		projectXYZ: {ID: projectXYZ, OrganisationID: orgXYZ},
	}
}

func jwtIdentity(userID, orgID string) *auth.Identity {
	return &auth.Identity{UserID: userID, OrgID: orgID, AuthMethod: auth.AuthMethodJWT}
}

func tokenIdentity(userID, orgID string, scopes []string, resourceIDs []string) *auth.Identity {
	return &auth.Identity{
		UserID:      userID,
		OrgID:       orgID,
		AuthMethod:  auth.AuthMethodAPIToken,
		TokenScopes: scopes,
		ResourceIDs: resourceIDs,
	}
}

var _ = Describe("PermissionService.Check", func() {

	Context("Identity validation", func() {
		It("should return Unauthenticated when no identity in context", func() {
			env := newTestEnv(GinkgoT(), defaultProjects())
			err := env.permService.Check(context.Background(), projectABC, auth.ResourceStacks, "", auth.ActionList)
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(errors.ErrorUnauthenticated))
		})
	})

	Context("API token scope enforcement", func() {
		var env *testEnv

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultProjects())
			// User is a developer. Project membership role also means org membership.
			Expect(env.policyMgr.AddGroupingPolicy(userDev, string(models.DeveloperRole), projectABC)).To(Succeed())
			Expect(env.policyMgr.AddGroupingPolicy(userDev, string(models.OrgMemberRole), orgABC)).To(Succeed())
		})

		DescribeTable("scope checks for user with developer role under org 'OrgABC' in project 'ProjectABC'",
			func(tc tokenPermCheck) {
				ctx := ctxWithIdentity(tokenIdentity(userDev, orgABC, tc.scopes, tc.resourceIDs))
				err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
				if tc.shouldAllow {
					Expect(err).To(BeNil())
				} else {
					Expect(err).ToNot(BeNil())
				}
			},
			Entry("matching scope", tokenPermCheck{
				scopes: []string{"stacks:*"}, domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("scope doesnt cover resource", tokenPermCheck{
				scopes: []string{"secrets:*"}, domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: false,
			}),
			Entry("full access scope", tokenPermCheck{
				scopes: []string{"*:*"}, domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("parent scope addons:* - read", tokenPermCheck{
				scopes: []string{"addons:*"}, domain: projectABC, resource: auth.ResourceAddonsPostgres, resourceID: "addon-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("parent scope addons:* - create", tokenPermCheck{
				scopes: []string{"addons:*"}, domain: projectABC, resource: auth.ResourceAddonsPostgres, resourceID: "addon-1", action: auth.ActionCreate, shouldAllow: true,
			}),
			Entry("specific action mismatch", tokenPermCheck{
				scopes: []string{"stacks:read"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionWrite, shouldAllow: false,
			}),
			Entry("resourceID restriction matching", tokenPermCheck{
				scopes: []string{"stacks:*"}, resourceIDs: []string{"s-1"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("resourceID restriction matching but scope action mismatch", tokenPermCheck{
				scopes: []string{"stacks:read"}, resourceIDs: []string{"s-1"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionWrite, shouldAllow: false,
			}),
			Entry("resourceID restriction not matching", tokenPermCheck{
				scopes: []string{"stacks:*"}, resourceIDs: []string{"s-1"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-2", action: auth.ActionRead, shouldAllow: false,
			}),
			Entry("empty resourceIDs no restriction", tokenPermCheck{
				scopes: []string{"stacks:*"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "any-id", action: auth.ActionRead, shouldAllow: true,
			}),
		)
	})

	Context("Developer role", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultProjects())
			Expect(env.policyMgr.AddGroupingPolicy(userDev, string(models.DeveloperRole), projectABC)).To(Succeed())
			Expect(env.policyMgr.AddGroupingPolicy(userDev, string(models.OrgMemberRole), orgABC)).To(Succeed())
			ctx = ctxWithIdentity(jwtIdentity(userDev, orgABC))
		})

		DescribeTable("project-scoped resources",
			func(tc permCheck) {
				err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
				if tc.shouldAllow {
					Expect(err).To(BeNil())
				} else {
					Expect(err).ToNot(BeNil())
				}
			},
			// Project resources — allowed
			Entry("list stacks in own project", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("create stack in own project", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionCreate, shouldAllow: true,
			}),
			Entry("read specific stack", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "stack-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("write specific stack", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "stack-1", action: auth.ActionWrite, shouldAllow: true,
			}),
			Entry("delete specific stack", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "stack-1", action: auth.ActionDelete, shouldAllow: true,
			}),
			Entry("list secrets", permCheck{
				domain: projectABC, resource: auth.ResourceSecrets, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("create secret", permCheck{
				domain: projectABC, resource: auth.ResourceSecrets, action: auth.ActionCreate, shouldAllow: true,
			}),
			Entry("read volume", permCheck{
				domain: projectABC, resource: auth.ResourceVolumes, resourceID: "vol-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("write volume", permCheck{
				domain: projectABC, resource: auth.ResourceVolumes, resourceID: "vol-1", action: auth.ActionWrite, shouldAllow: true,
			}),
			Entry("delete volume", permCheck{
				domain: projectABC, resource: auth.ResourceVolumes, resourceID: "vol-1", action: auth.ActionDelete, shouldAllow: true,
			}),
			Entry("read addon", permCheck{
				domain: projectABC, resource: auth.ResourceAddonsPostgres, resourceID: "addon-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("write addon", permCheck{
				domain: projectABC, resource: auth.ResourceAddonsPostgres, resourceID: "addon-1", action: auth.ActionWrite, shouldAllow: true,
			}),
			Entry("list object stores", permCheck{
				domain: projectABC, resource: auth.ResourceObjectStores, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("delete object store", permCheck{
				domain: projectABC, resource: auth.ResourceObjectStores, resourceID: "os-1", action: auth.ActionDelete, shouldAllow: true,
			}),
			Entry("create workspace user", permCheck{
				domain: projectABC, resource: auth.ResourceWorkspaceUsers, action: auth.ActionCreate, shouldAllow: true,
			}),
			Entry("write workspace user", permCheck{
				domain: projectABC, resource: auth.ResourceWorkspaceUsers, resourceID: "wu-1", action: auth.ActionWrite, shouldAllow: true,
			}),

			// Org-scoped via OrgMember — read only
			Entry("list clusters via OrgMember", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read cluster via OrgMember", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, resourceID: "c-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("cannot create cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionCreate, shouldAllow: false,
			}),
			Entry("cannot delete cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, resourceID: "c-1", action: auth.ActionDelete, shouldAllow: false,
			}),
			Entry("list image registries", permCheck{
				domain: orgABC, resource: auth.ResourceImageRegistries, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read image registry", permCheck{
				domain: orgABC, resource: auth.ResourceImageRegistries, resourceID: "reg-1", action: auth.ActionRead, shouldAllow: true,
			}),

			// Cross-project isolation
			Entry("cannot access other projects stacks", permCheck{
				domain: projectXYZ, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: false,
			}),
			Entry("cannot read other projects secret", permCheck{
				domain: projectXYZ, resource: auth.ResourceSecrets, resourceID: "sec-1", action: auth.ActionRead, shouldAllow: false,
			}),
		)
	})

	Context("Viewer role", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultProjects())
			Expect(env.policyMgr.AddGroupingPolicy(userView, string(models.ViewerRole), projectABC)).To(Succeed())
			Expect(env.policyMgr.AddGroupingPolicy(userView, string(models.OrgMemberRole), orgABC)).To(Succeed())
			ctx = ctxWithIdentity(jwtIdentity(userView, orgABC))
		})

		DescribeTable("read-only access",
			func(tc permCheck) {
				err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
				if tc.shouldAllow {
					Expect(err).To(BeNil())
				} else {
					Expect(err).ToNot(BeNil())
				}
			},
			// Read access — allowed
			Entry("list stacks", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read stack", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "stack-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("list secrets", permCheck{
				domain: projectABC, resource: auth.ResourceSecrets, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read secret", permCheck{
				domain: projectABC, resource: auth.ResourceSecrets, resourceID: "sec-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("list volumes", permCheck{
				domain: projectABC, resource: auth.ResourceVolumes, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read volume", permCheck{
				domain: projectABC, resource: auth.ResourceVolumes, resourceID: "vol-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("read addon", permCheck{
				domain: projectABC, resource: auth.ResourceAddonsPostgres, resourceID: "addon-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("list object stores", permCheck{
				domain: projectABC, resource: auth.ResourceObjectStores, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read object store", permCheck{
				domain: projectABC, resource: auth.ResourceObjectStores, resourceID: "os-1", action: auth.ActionRead, shouldAllow: true,
			}),

			// Write access — denied
			Entry("cannot create stack", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionCreate, shouldAllow: false,
			}),
			Entry("cannot write stack", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "stack-1", action: auth.ActionWrite, shouldAllow: false,
			}),
			Entry("cannot delete stack", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "stack-1", action: auth.ActionDelete, shouldAllow: false,
			}),
			Entry("cannot create secret", permCheck{
				domain: projectABC, resource: auth.ResourceSecrets, action: auth.ActionCreate, shouldAllow: false,
			}),
			Entry("cannot delete secret", permCheck{
				domain: projectABC, resource: auth.ResourceSecrets, resourceID: "sec-1", action: auth.ActionDelete, shouldAllow: false,
			}),
			Entry("cannot delete addon", permCheck{
				domain: projectABC, resource: auth.ResourceAddonsPostgres, resourceID: "addon-1", action: auth.ActionDelete, shouldAllow: false,
			}),

			// Org-scoped via OrgMember
			Entry("list clusters via OrgMember", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read cluster via OrgMember", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, resourceID: "c-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("cannot create cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionCreate, shouldAllow: false,
			}),
			Entry("cannot delete cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, resourceID: "c-1", action: auth.ActionDelete, shouldAllow: false,
			}),
			Entry("list image registries", permCheck{
				domain: orgABC, resource: auth.ResourceImageRegistries, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read image registry", permCheck{
				domain: orgABC, resource: auth.ResourceImageRegistries, resourceID: "reg-1", action: auth.ActionRead, shouldAllow: true,
			}),
		)
	})

	Context("OrgMember role (no project membership)", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultProjects())
			Expect(env.policyMgr.AddGroupingPolicy(userMem, string(models.OrgMemberRole), orgABC)).To(Succeed())
			ctx = ctxWithIdentity(jwtIdentity(userMem, orgABC))
		})

		DescribeTable("org-scoped access only",
			func(tc permCheck) {
				err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
				if tc.shouldAllow {
					Expect(err).To(BeNil())
				} else {
					Expect(err).ToNot(BeNil())
				}
			},
			// Org-scoped read — allowed
			Entry("list clusters", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, resourceID: "c-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("list image registries", permCheck{
				domain: orgABC, resource: auth.ResourceImageRegistries, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read image registry", permCheck{
				domain: orgABC, resource: auth.ResourceImageRegistries, resourceID: "reg-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("read org", permCheck{
				domain: orgABC, resource: auth.ResourceOrgs, resourceID: orgABC, action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("list projects", permCheck{
				domain: orgABC, resource: auth.ResourceProjects, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read project", permCheck{
				domain: orgABC, resource: auth.ResourceProjects, resourceID: "project-1", action: auth.ActionRead, shouldAllow: true,
			}),

			// Org-scoped write — denied
			Entry("cannot create cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionCreate, shouldAllow: false,
			}),
			Entry("cannot delete cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, resourceID: "c-1", action: auth.ActionDelete, shouldAllow: false,
			}),
			Entry("cannot write org", permCheck{
				domain: orgABC, resource: auth.ResourceOrgs, resourceID: orgABC, action: auth.ActionWrite, shouldAllow: false,
			}),

			// Project-scoped — denied
			Entry("cannot list stacks", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: false,
			}),
			Entry("cannot list secrets", permCheck{
				domain: projectXYZ, resource: auth.ResourceSecrets, action: auth.ActionList, shouldAllow: false,
			}),
		)
	})

	Context("OrgAdmin fallback", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultProjects())
			Expect(env.policyMgr.AddGroupingPolicy(userAdmin, string(models.OrgAdminRole), orgABC)).To(Succeed())
			ctx = ctxWithIdentity(jwtIdentity(userAdmin, orgABC))
		})

		DescribeTable("access via OrgAdmin",
			func(tc permCheck) {
				err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
				if tc.shouldAllow {
					Expect(err).To(BeNil())
				} else {
					Expect(err).ToNot(BeNil())
				}
			},
			// Project resources via fallback (project-abc in org-abc)
			Entry("list stacks in project", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("read stack in project", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "stack-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("delete secret in project", permCheck{
				domain: projectABC, resource: auth.ResourceSecrets, resourceID: "sec-1", action: auth.ActionDelete, shouldAllow: true,
			}),
			Entry("create addon in project", permCheck{
				domain: projectABC, resource: auth.ResourceAddonsPostgres, action: auth.ActionCreate, shouldAllow: true,
			}),

			// Org resources directly
			Entry("list clusters", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionList, shouldAllow: true,
			}),
			Entry("create cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionCreate, shouldAllow: true,
			}),
			Entry("delete cluster", permCheck{
				domain: orgABC, resource: auth.ResourceClusters, resourceID: "c-1", action: auth.ActionDelete, shouldAllow: true,
			}),

			// Other org's project — denied
			Entry("cannot access other orgs project", permCheck{
				domain: projectXYZ, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: false,
			}),
			Entry("cannot read other orgs secret", permCheck{
				domain: projectXYZ, resource: auth.ResourceSecrets, resourceID: "sec-1", action: auth.ActionRead, shouldAllow: false,
			}),

			// Unknown project — denied
			Entry("unknown project domain", permCheck{
				domain: "nonexistent", resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: false,
			}),
		)
	})

	Context("Cross-project isolation", func() {
		It("should isolate users to their own projects", func() {
			projects := map[string]*models.Project{
				projectABC: {ID: projectABC, OrganisationID: orgABC},
				projectXYZ: {ID: projectXYZ, OrganisationID: orgABC},
			}
			env := newTestEnv(GinkgoT(), projects)

			user1 := "user-1"
			user2 := "user-2"
			Expect(env.policyMgr.AddGroupingPolicy(user1, string(models.DeveloperRole), projectABC)).To(Succeed())
			Expect(env.policyMgr.AddGroupingPolicy(user1, string(models.OrgMemberRole), orgABC)).To(Succeed())
			Expect(env.policyMgr.AddGroupingPolicy(user2, string(models.DeveloperRole), projectXYZ)).To(Succeed())
			Expect(env.policyMgr.AddGroupingPolicy(user2, string(models.OrgMemberRole), orgABC)).To(Succeed())

			ctx1 := ctxWithIdentity(jwtIdentity(user1, orgABC))
			ctx2 := ctxWithIdentity(jwtIdentity(user2, orgABC))

			Expect(env.permService.Check(ctx1, projectABC, auth.ResourceStacks, "", auth.ActionList)).To(BeNil())
			Expect(env.permService.Check(ctx1, projectXYZ, auth.ResourceStacks, "", auth.ActionList)).ToNot(BeNil())
			Expect(env.permService.Check(ctx2, projectXYZ, auth.ResourceStacks, "", auth.ActionList)).To(BeNil())
			Expect(env.permService.Check(ctx2, projectABC, auth.ResourceStacks, "", auth.ActionList)).ToNot(BeNil())
		})
	})

	Context("Multi-role user", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			projects := map[string]*models.Project{
				projectABC: {ID: projectABC, OrganisationID: orgABC},
				projectXYZ: {ID: projectXYZ, OrganisationID: orgABC},
			}
			env = newTestEnv(GinkgoT(), projects)
			user := "user-multi"
			Expect(env.policyMgr.AddGroupingPolicy(user, string(models.DeveloperRole), projectABC)).To(Succeed())
			Expect(env.policyMgr.AddGroupingPolicy(user, string(models.ViewerRole), projectXYZ)).To(Succeed())
			Expect(env.policyMgr.AddGroupingPolicy(user, string(models.OrgMemberRole), orgABC)).To(Succeed())
			ctx = ctxWithIdentity(jwtIdentity(user, orgABC))
		})

		DescribeTable("permissions depend on project role",
			func(tc permCheck) {
				err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
				if tc.shouldAllow {
					Expect(err).To(BeNil())
				} else {
					Expect(err).ToNot(BeNil())
				}
			},
			Entry("write in Developer project", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionWrite, shouldAllow: true,
			}),
			Entry("delete in Developer project", permCheck{
				domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionDelete, shouldAllow: true,
			}),
			Entry("read in Viewer project", permCheck{
				domain: projectXYZ, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionRead, shouldAllow: true,
			}),
			Entry("cannot write in Viewer project", permCheck{
				domain: projectXYZ, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionWrite, shouldAllow: false,
			}),
			Entry("cannot delete in Viewer project", permCheck{
				domain: projectXYZ, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionDelete, shouldAllow: false,
			}),
			Entry("cannot create in Viewer project", permCheck{
				domain: projectXYZ, resource: auth.ResourceStacks, action: auth.ActionCreate, shouldAllow: false,
			}),
		)
	})

	Context("API token + Casbin combined (effective permissions = min of both)", func() {

		Context("Viewer with token", func() {
			var env *testEnv

			BeforeEach(func() {
				env = newTestEnv(GinkgoT(), defaultProjects())
				Expect(env.policyMgr.AddGroupingPolicy(userView, string(models.ViewerRole), projectABC)).To(Succeed())
				Expect(env.policyMgr.AddGroupingPolicy(userView, string(models.OrgMemberRole), orgABC)).To(Succeed())
			})

			DescribeTable("token cannot expand Viewer permissions",
				func(tc tokenPermCheck) {
					ctx := ctxWithIdentity(tokenIdentity(userView, orgABC, tc.scopes, nil))
					err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
					if tc.shouldAllow {
						Expect(err).To(BeNil())
					} else {
						Expect(err).ToNot(BeNil())
					}
				},
				Entry("full token viewer can read", tokenPermCheck{
					scopes: []string{"*:*"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionRead, shouldAllow: true,
				}),
				Entry("full token viewer cannot write", tokenPermCheck{
					scopes: []string{"*:*"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionWrite, shouldAllow: false,
				}),
				Entry("full token viewer cannot delete", tokenPermCheck{
					scopes: []string{"*:*"}, domain: projectABC, resource: auth.ResourceSecrets, resourceID: "s-1", action: auth.ActionDelete, shouldAllow: false,
				}),
				Entry("full token viewer cannot create", tokenPermCheck{
					scopes: []string{"*:*"}, domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionCreate, shouldAllow: false,
				}),
				Entry("stacks token viewer reads stack", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionRead, shouldAllow: true,
				}),
				Entry("stacks token blocks secret read", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: projectABC, resource: auth.ResourceSecrets, resourceID: "s-1", action: auth.ActionRead, shouldAllow: false,
				}),
				Entry("read token viewer reads", tokenPermCheck{
					scopes: []string{"stacks:read"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionRead, shouldAllow: true,
				}),
				Entry("list token viewer lists", tokenPermCheck{
					scopes: []string{"stacks:list"}, domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: true,
				}),
				Entry("list token viewer can list clusters - specific token", tokenPermCheck{
					scopes: []string{"clusters:list"}, domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionList, shouldAllow: true,
				}),
				Entry("list token viewer can list clusters - full token", tokenPermCheck{
					scopes: []string{"*:*"}, domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionList, shouldAllow: true,
				}),
			)
		})

		Context("Developer with restricted token", func() {
			var env *testEnv

			BeforeEach(func() {
				env = newTestEnv(GinkgoT(), defaultProjects())
				Expect(env.policyMgr.AddGroupingPolicy(userDev, string(models.DeveloperRole), projectABC)).To(Succeed())
				Expect(env.policyMgr.AddGroupingPolicy(userDev, string(models.OrgMemberRole), orgABC)).To(Succeed())
			})

			DescribeTable("token restricts Developer permissions",
				func(tc tokenPermCheck) {
					ctx := ctxWithIdentity(tokenIdentity(userDev, orgABC, tc.scopes, nil))
					err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
					if tc.shouldAllow {
						Expect(err).To(BeNil())
					} else {
						Expect(err).ToNot(BeNil())
					}
				},
				Entry("read-only token blocks developer write", tokenPermCheck{
					scopes: []string{"stacks:read"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionWrite, shouldAllow: false,
				}),
				Entry("secrets token blocks developer stacks", tokenPermCheck{
					scopes: []string{"secrets:*"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionRead, shouldAllow: false,
				}),
				Entry("stacks token allows developer stacks", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: projectABC, resource: auth.ResourceStacks, resourceID: "s-1", action: auth.ActionWrite, shouldAllow: true,
				}),
			)
		})

		Context("OrgAdmin with token", func() {
			var env *testEnv

			BeforeEach(func() {
				env = newTestEnv(GinkgoT(), defaultProjects())
				Expect(env.policyMgr.AddGroupingPolicy(userAdmin, string(models.OrgAdminRole), orgABC)).To(Succeed())
			})

			DescribeTable("token restricts OrgAdmin permissions",
				func(tc tokenPermCheck) {
					ctx := ctxWithIdentity(tokenIdentity(userAdmin, orgABC, tc.scopes, nil))
					err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
					if tc.shouldAllow {
						Expect(err).To(BeNil())
					} else {
						Expect(err).ToNot(BeNil())
					}
				},
				Entry("scoped token accesses allowed resource", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: true,
				}),
				Entry("scoped token blocked on wrong resource", tokenPermCheck{
					scopes: []string{"secrets:*"}, domain: projectABC, resource: auth.ResourceStacks, action: auth.ActionList, shouldAllow: false,
				}),
				Entry("clusters token accesses org resource", tokenPermCheck{
					scopes: []string{"clusters:*"}, domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionList, shouldAllow: true,
				}),
				Entry("stacks token blocks on clusters", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: orgABC, resource: auth.ResourceClusters, action: auth.ActionList, shouldAllow: false,
				}),
			)
		})
	})

	Context("Edge cases", func() {
		var env *testEnv

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultProjects())
			Expect(env.policyMgr.AddGroupingPolicy(userAdmin, string(models.OrgAdminRole), orgABC)).To(Succeed())
		})

		It("should deny empty domain even for OrgAdmin", func() {
			ctx := ctxWithIdentity(jwtIdentity(userAdmin, orgABC))
			err := env.permService.Check(ctx, "", auth.ResourceStacks, "", auth.ActionList)
			Expect(err).ToNot(BeNil())
		})

		It("should allow OrgAdmin with empty resource via wildcard", func() {
			ctx := ctxWithIdentity(jwtIdentity(userAdmin, orgABC))
			err := env.permService.Check(ctx, orgABC, "", "", auth.ActionList)
			Expect(err).To(BeNil())
		})
	})
})
