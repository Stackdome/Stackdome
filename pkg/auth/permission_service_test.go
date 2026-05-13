package auth

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func TestPermissionService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PermissionService Suite")
}

const (
	teamABC   = "team-abc"
	teamXYZ   = "team-xyz"
	orgABC    = "org-abc"
	orgXYZ    = "org-xyz"
	userDev   = "user-dev"
	userView  = "user-view"
	userMem   = "user-mem"
	userAdmin = "user-admin"
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

func defaultTeams() map[string]*models.Team {
	return map[string]*models.Team{
		teamABC: {ID: teamABC, OrganisationID: orgABC},
		teamXYZ: {ID: teamXYZ, OrganisationID: orgXYZ},
	}
}

func jwtIdentity(userID, orgID string) *Identity {
	return &Identity{UserID: userID, OrgID: orgID, AuthMethod: AuthMethodJWT}
}

func tokenIdentity(userID, orgID string, scopes []string, resourceIDs []string) *Identity {
	return &Identity{
		UserID:      userID,
		OrgID:       orgID,
		AuthMethod:  AuthMethodAPIToken,
		TokenScopes: scopes,
		ResourceIDs: resourceIDs,
	}
}

var _ = Describe("PermissionService.Check", func() {

	Context("Identity validation", func() {
		It("should return Unauthenticated when no identity in context", func() {
			env := newTestEnv(GinkgoT(), defaultTeams())
			err := env.permService.Check(context.Background(), teamABC, ResourceStacks, "", ActionList)
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(errors.ErrorUnauthenticated))
		})
	})

	Context("API token scope enforcement", func() {
		var env *testEnv

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultTeams())
			// User is a developer. Team membership role also means org membership.
			env.policyMgr.AddGroupingPolicy(userDev, string(models.DeveloperRole), teamABC)
			env.policyMgr.AddGroupingPolicy(userDev, string(models.OrgMemberRole), orgABC)
		})

		DescribeTable("scope checks for user with developer role under org 'OrgABC' in team 'TeamABC'",
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
				scopes: []string{"stacks:*"}, domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: true,
			}),
			Entry("scope doesnt cover resource", tokenPermCheck{
				scopes: []string{"secrets:*"}, domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: false,
			}),
			Entry("full access scope", tokenPermCheck{
				scopes: []string{"*:*"}, domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: true,
			}),
			Entry("parent scope addons:* - read", tokenPermCheck{
				scopes: []string{"addons:*"}, domain: teamABC, resource: ResourceAddonsPostgres, resourceID: "addon-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("parent scope addons:* - create", tokenPermCheck{
				scopes: []string{"addons:*"}, domain: teamABC, resource: ResourceAddonsPostgres, resourceID: "addon-1", action: ActionCreate, shouldAllow: true,
			}),
			Entry("specific action mismatch", tokenPermCheck{
				scopes: []string{"stacks:read"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionWrite, shouldAllow: false,
			}),
			Entry("resourceID restriction matching", tokenPermCheck{
				scopes: []string{"stacks:*"}, resourceIDs: []string{"s-1"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("resourceID restriction matching but scope action mismatch", tokenPermCheck{
				scopes: []string{"stacks:read"}, resourceIDs: []string{"s-1"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionWrite, shouldAllow: false,
			}),
			Entry("resourceID restriction not matching", tokenPermCheck{
				scopes: []string{"stacks:*"}, resourceIDs: []string{"s-1"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-2", action: ActionRead, shouldAllow: false,
			}),
			Entry("empty resourceIDs no restriction", tokenPermCheck{
				scopes: []string{"stacks:*"}, domain: teamABC, resource: ResourceStacks, resourceID: "any-id", action: ActionRead, shouldAllow: true,
			}),
		)
	})

	Context("Developer role", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultTeams())
			env.policyMgr.AddGroupingPolicy(userDev, string(models.DeveloperRole), teamABC)
			env.policyMgr.AddGroupingPolicy(userDev, string(models.OrgMemberRole), orgABC)
			ctx = ctxWithIdentity(jwtIdentity(userDev, orgABC))
		})

		DescribeTable("team-scoped resources",
			func(tc permCheck) {
				err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
				if tc.shouldAllow {
					Expect(err).To(BeNil())
				} else {
					Expect(err).ToNot(BeNil())
				}
			},
			// Team resources — allowed
			Entry("list stacks in own team", permCheck{
				domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: true,
			}),
			Entry("create stack in own team", permCheck{
				domain: teamABC, resource: ResourceStacks, action: ActionCreate, shouldAllow: true,
			}),
			Entry("read specific stack", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "stack-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("write specific stack", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "stack-1", action: ActionWrite, shouldAllow: true,
			}),
			Entry("delete specific stack", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "stack-1", action: ActionDelete, shouldAllow: true,
			}),
			Entry("list secrets", permCheck{
				domain: teamABC, resource: ResourceSecrets, action: ActionList, shouldAllow: true,
			}),
			Entry("create secret", permCheck{
				domain: teamABC, resource: ResourceSecrets, action: ActionCreate, shouldAllow: true,
			}),
			Entry("read volume", permCheck{
				domain: teamABC, resource: ResourceVolumes, resourceID: "vol-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("write volume", permCheck{
				domain: teamABC, resource: ResourceVolumes, resourceID: "vol-1", action: ActionWrite, shouldAllow: true,
			}),
			Entry("delete volume", permCheck{
				domain: teamABC, resource: ResourceVolumes, resourceID: "vol-1", action: ActionDelete, shouldAllow: true,
			}),
			Entry("read addon", permCheck{
				domain: teamABC, resource: ResourceAddonsPostgres, resourceID: "addon-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("write addon", permCheck{
				domain: teamABC, resource: ResourceAddonsPostgres, resourceID: "addon-1", action: ActionWrite, shouldAllow: true,
			}),
			Entry("list object stores", permCheck{
				domain: teamABC, resource: ResourceObjectStores, action: ActionList, shouldAllow: true,
			}),
			Entry("delete object store", permCheck{
				domain: teamABC, resource: ResourceObjectStores, resourceID: "os-1", action: ActionDelete, shouldAllow: true,
			}),
			Entry("create workspace user", permCheck{
				domain: teamABC, resource: ResourceWorkspaceUsers, action: ActionCreate, shouldAllow: true,
			}),
			Entry("write workspace user", permCheck{
				domain: teamABC, resource: ResourceWorkspaceUsers, resourceID: "wu-1", action: ActionWrite, shouldAllow: true,
			}),

			// Org-scoped via OrgMember — read only
			Entry("list clusters via OrgMember", permCheck{
				domain: orgABC, resource: ResourceClusters, action: ActionList, shouldAllow: true,
			}),
			Entry("read cluster via OrgMember", permCheck{
				domain: orgABC, resource: ResourceClusters, resourceID: "c-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("cannot create cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, action: ActionCreate, shouldAllow: false,
			}),
			Entry("cannot delete cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, resourceID: "c-1", action: ActionDelete, shouldAllow: false,
			}),
			Entry("list image registries", permCheck{
				domain: orgABC, resource: ResourceImageRegistries, action: ActionList, shouldAllow: true,
			}),
			Entry("read image registry", permCheck{
				domain: orgABC, resource: ResourceImageRegistries, resourceID: "reg-1", action: ActionRead, shouldAllow: true,
			}),

			// Cross-team isolation
			Entry("cannot access other teams stacks", permCheck{
				domain: teamXYZ, resource: ResourceStacks, action: ActionList, shouldAllow: false,
			}),
			Entry("cannot read other teams secret", permCheck{
				domain: teamXYZ, resource: ResourceSecrets, resourceID: "sec-1", action: ActionRead, shouldAllow: false,
			}),
		)
	})

	Context("Viewer role", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultTeams())
			env.policyMgr.AddGroupingPolicy(userView, string(models.ViewerRole), teamABC)
			env.policyMgr.AddGroupingPolicy(userView, string(models.OrgMemberRole), orgABC)
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
				domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: true,
			}),
			Entry("read stack", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "stack-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("list secrets", permCheck{
				domain: teamABC, resource: ResourceSecrets, action: ActionList, shouldAllow: true,
			}),
			Entry("read secret", permCheck{
				domain: teamABC, resource: ResourceSecrets, resourceID: "sec-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("list volumes", permCheck{
				domain: teamABC, resource: ResourceVolumes, action: ActionList, shouldAllow: true,
			}),
			Entry("read volume", permCheck{
				domain: teamABC, resource: ResourceVolumes, resourceID: "vol-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("read addon", permCheck{
				domain: teamABC, resource: ResourceAddonsPostgres, resourceID: "addon-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("list object stores", permCheck{
				domain: teamABC, resource: ResourceObjectStores, action: ActionList, shouldAllow: true,
			}),
			Entry("read object store", permCheck{
				domain: teamABC, resource: ResourceObjectStores, resourceID: "os-1", action: ActionRead, shouldAllow: true,
			}),

			// Write access — denied
			Entry("cannot create stack", permCheck{
				domain: teamABC, resource: ResourceStacks, action: ActionCreate, shouldAllow: false,
			}),
			Entry("cannot write stack", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "stack-1", action: ActionWrite, shouldAllow: false,
			}),
			Entry("cannot delete stack", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "stack-1", action: ActionDelete, shouldAllow: false,
			}),
			Entry("cannot create secret", permCheck{
				domain: teamABC, resource: ResourceSecrets, action: ActionCreate, shouldAllow: false,
			}),
			Entry("cannot delete secret", permCheck{
				domain: teamABC, resource: ResourceSecrets, resourceID: "sec-1", action: ActionDelete, shouldAllow: false,
			}),
			Entry("cannot delete addon", permCheck{
				domain: teamABC, resource: ResourceAddonsPostgres, resourceID: "addon-1", action: ActionDelete, shouldAllow: false,
			}),

			// Org-scoped via OrgMember
			Entry("list clusters via OrgMember", permCheck{
				domain: orgABC, resource: ResourceClusters, action: ActionList, shouldAllow: true,
			}),
			Entry("read cluster via OrgMember", permCheck{
				domain: orgABC, resource: ResourceClusters, resourceID: "c-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("cannot create cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, action: ActionCreate, shouldAllow: false,
			}),
			Entry("cannot delete cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, resourceID: "c-1", action: ActionDelete, shouldAllow: false,
			}),
			Entry("list image registries", permCheck{
				domain: orgABC, resource: ResourceImageRegistries, action: ActionList, shouldAllow: true,
			}),
			Entry("read image registry", permCheck{
				domain: orgABC, resource: ResourceImageRegistries, resourceID: "reg-1", action: ActionRead, shouldAllow: true,
			}),
		)
	})

	Context("OrgMember role (no team membership)", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultTeams())
			env.policyMgr.AddGroupingPolicy(userMem, string(models.OrgMemberRole), orgABC)
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
				domain: orgABC, resource: ResourceClusters, action: ActionList, shouldAllow: true,
			}),
			Entry("read cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, resourceID: "c-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("list image registries", permCheck{
				domain: orgABC, resource: ResourceImageRegistries, action: ActionList, shouldAllow: true,
			}),
			Entry("read image registry", permCheck{
				domain: orgABC, resource: ResourceImageRegistries, resourceID: "reg-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("read org", permCheck{
				domain: orgABC, resource: ResourceOrgs, resourceID: orgABC, action: ActionRead, shouldAllow: true,
			}),
			Entry("list teams", permCheck{
				domain: orgABC, resource: ResourceTeams, action: ActionList, shouldAllow: true,
			}),
			Entry("read team", permCheck{
				domain: orgABC, resource: ResourceTeams, resourceID: "team-1", action: ActionRead, shouldAllow: true,
			}),

			// Org-scoped write — denied
			Entry("cannot create cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, action: ActionCreate, shouldAllow: false,
			}),
			Entry("cannot delete cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, resourceID: "c-1", action: ActionDelete, shouldAllow: false,
			}),
			Entry("cannot write org", permCheck{
				domain: orgABC, resource: ResourceOrgs, resourceID: orgABC, action: ActionWrite, shouldAllow: false,
			}),

			// Team-scoped — denied
			Entry("cannot list stacks", permCheck{
				domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: false,
			}),
			Entry("cannot list secrets", permCheck{
				domain: teamXYZ, resource: ResourceSecrets, action: ActionList, shouldAllow: false,
			}),
		)
	})

	Context("OrgAdmin fallback", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultTeams())
			env.policyMgr.AddGroupingPolicy(userAdmin, string(models.OrgAdminRole), orgABC)
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
			// Team resources via fallback (team-abc in org-abc)
			Entry("list stacks in team", permCheck{
				domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: true,
			}),
			Entry("read stack in team", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "stack-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("delete secret in team", permCheck{
				domain: teamABC, resource: ResourceSecrets, resourceID: "sec-1", action: ActionDelete, shouldAllow: true,
			}),
			Entry("create addon in team", permCheck{
				domain: teamABC, resource: ResourceAddonsPostgres, action: ActionCreate, shouldAllow: true,
			}),

			// Org resources directly
			Entry("list clusters", permCheck{
				domain: orgABC, resource: ResourceClusters, action: ActionList, shouldAllow: true,
			}),
			Entry("create cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, action: ActionCreate, shouldAllow: true,
			}),
			Entry("delete cluster", permCheck{
				domain: orgABC, resource: ResourceClusters, resourceID: "c-1", action: ActionDelete, shouldAllow: true,
			}),

			// Other org's team — denied
			Entry("cannot access other orgs team", permCheck{
				domain: teamXYZ, resource: ResourceStacks, action: ActionList, shouldAllow: false,
			}),
			Entry("cannot read other orgs secret", permCheck{
				domain: teamXYZ, resource: ResourceSecrets, resourceID: "sec-1", action: ActionRead, shouldAllow: false,
			}),

			// Unknown team — denied
			Entry("unknown team domain", permCheck{
				domain: "nonexistent", resource: ResourceStacks, action: ActionList, shouldAllow: false,
			}),
		)
	})

	Context("Cross-team isolation", func() {
		It("should isolate users to their own teams", func() {
			teams := map[string]*models.Team{
				teamABC: {ID: teamABC, OrganisationID: orgABC},
				teamXYZ: {ID: teamXYZ, OrganisationID: orgABC},
			}
			env := newTestEnv(GinkgoT(), teams)

			user1 := "user-1"
			user2 := "user-2"
			env.policyMgr.AddGroupingPolicy(user1, string(models.DeveloperRole), teamABC)
			env.policyMgr.AddGroupingPolicy(user1, string(models.OrgMemberRole), orgABC)
			env.policyMgr.AddGroupingPolicy(user2, string(models.DeveloperRole), teamXYZ)
			env.policyMgr.AddGroupingPolicy(user2, string(models.OrgMemberRole), orgABC)

			ctx1 := ctxWithIdentity(jwtIdentity(user1, orgABC))
			ctx2 := ctxWithIdentity(jwtIdentity(user2, orgABC))

			Expect(env.permService.Check(ctx1, teamABC, ResourceStacks, "", ActionList)).To(BeNil())
			Expect(env.permService.Check(ctx1, teamXYZ, ResourceStacks, "", ActionList)).ToNot(BeNil())
			Expect(env.permService.Check(ctx2, teamXYZ, ResourceStacks, "", ActionList)).To(BeNil())
			Expect(env.permService.Check(ctx2, teamABC, ResourceStacks, "", ActionList)).ToNot(BeNil())
		})
	})

	Context("Multi-role user", func() {
		var (
			env *testEnv
			ctx context.Context
		)

		BeforeEach(func() {
			teams := map[string]*models.Team{
				teamABC: {ID: teamABC, OrganisationID: orgABC},
				teamXYZ: {ID: teamXYZ, OrganisationID: orgABC},
			}
			env = newTestEnv(GinkgoT(), teams)
			user := "user-multi"
			env.policyMgr.AddGroupingPolicy(user, string(models.DeveloperRole), teamABC)
			env.policyMgr.AddGroupingPolicy(user, string(models.ViewerRole), teamXYZ)
			env.policyMgr.AddGroupingPolicy(user, string(models.OrgMemberRole), orgABC)
			ctx = ctxWithIdentity(jwtIdentity(user, orgABC))
		})

		DescribeTable("permissions depend on team role",
			func(tc permCheck) {
				err := env.permService.Check(ctx, tc.domain, tc.resource, tc.resourceID, tc.action)
				if tc.shouldAllow {
					Expect(err).To(BeNil())
				} else {
					Expect(err).ToNot(BeNil())
				}
			},
			Entry("write in Developer team", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionWrite, shouldAllow: true,
			}),
			Entry("delete in Developer team", permCheck{
				domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionDelete, shouldAllow: true,
			}),
			Entry("read in Viewer team", permCheck{
				domain: teamXYZ, resource: ResourceStacks, resourceID: "s-1", action: ActionRead, shouldAllow: true,
			}),
			Entry("cannot write in Viewer team", permCheck{
				domain: teamXYZ, resource: ResourceStacks, resourceID: "s-1", action: ActionWrite, shouldAllow: false,
			}),
			Entry("cannot delete in Viewer team", permCheck{
				domain: teamXYZ, resource: ResourceStacks, resourceID: "s-1", action: ActionDelete, shouldAllow: false,
			}),
			Entry("cannot create in Viewer team", permCheck{
				domain: teamXYZ, resource: ResourceStacks, action: ActionCreate, shouldAllow: false,
			}),
		)
	})

	Context("API token + Casbin combined (effective permissions = min of both)", func() {

		Context("Viewer with token", func() {
			var env *testEnv

			BeforeEach(func() {
				env = newTestEnv(GinkgoT(), defaultTeams())
				env.policyMgr.AddGroupingPolicy(userView, string(models.ViewerRole), teamABC)
				env.policyMgr.AddGroupingPolicy(userView, string(models.OrgMemberRole), orgABC)
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
					scopes: []string{"*:*"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionRead, shouldAllow: true,
				}),
				Entry("full token viewer cannot write", tokenPermCheck{
					scopes: []string{"*:*"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionWrite, shouldAllow: false,
				}),
				Entry("full token viewer cannot delete", tokenPermCheck{
					scopes: []string{"*:*"}, domain: teamABC, resource: ResourceSecrets, resourceID: "s-1", action: ActionDelete, shouldAllow: false,
				}),
				Entry("full token viewer cannot create", tokenPermCheck{
					scopes: []string{"*:*"}, domain: teamABC, resource: ResourceStacks, action: ActionCreate, shouldAllow: false,
				}),
				Entry("stacks token viewer reads stack", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionRead, shouldAllow: true,
				}),
				Entry("stacks token blocks secret read", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: teamABC, resource: ResourceSecrets, resourceID: "s-1", action: ActionRead, shouldAllow: false,
				}),
				Entry("read token viewer reads", tokenPermCheck{
					scopes: []string{"stacks:read"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionRead, shouldAllow: true,
				}),
				Entry("list token viewer lists", tokenPermCheck{
					scopes: []string{"stacks:list"}, domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: true,
				}),
				Entry("list token viewer can list clusters - specific token", tokenPermCheck{
					scopes: []string{"clusters:list"}, domain: orgABC, resource: ResourceClusters, action: ActionList, shouldAllow: true,
				}),
				Entry("list token viewer can list clusters - full token", tokenPermCheck{
					scopes: []string{"*:*"}, domain: orgABC, resource: ResourceClusters, action: ActionList, shouldAllow: true,
				}),
			)
		})

		Context("Developer with restricted token", func() {
			var env *testEnv

			BeforeEach(func() {
				env = newTestEnv(GinkgoT(), defaultTeams())
				env.policyMgr.AddGroupingPolicy(userDev, string(models.DeveloperRole), teamABC)
				env.policyMgr.AddGroupingPolicy(userDev, string(models.OrgMemberRole), orgABC)
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
					scopes: []string{"stacks:read"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionWrite, shouldAllow: false,
				}),
				Entry("secrets token blocks developer stacks", tokenPermCheck{
					scopes: []string{"secrets:*"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionRead, shouldAllow: false,
				}),
				Entry("stacks token allows developer stacks", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: teamABC, resource: ResourceStacks, resourceID: "s-1", action: ActionWrite, shouldAllow: true,
				}),
			)
		})

		Context("OrgAdmin with token", func() {
			var env *testEnv

			BeforeEach(func() {
				env = newTestEnv(GinkgoT(), defaultTeams())
				env.policyMgr.AddGroupingPolicy(userAdmin, string(models.OrgAdminRole), orgABC)
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
					scopes: []string{"stacks:*"}, domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: true,
				}),
				Entry("scoped token blocked on wrong resource", tokenPermCheck{
					scopes: []string{"secrets:*"}, domain: teamABC, resource: ResourceStacks, action: ActionList, shouldAllow: false,
				}),
				Entry("clusters token accesses org resource", tokenPermCheck{
					scopes: []string{"clusters:*"}, domain: orgABC, resource: ResourceClusters, action: ActionList, shouldAllow: true,
				}),
				Entry("stacks token blocks on clusters", tokenPermCheck{
					scopes: []string{"stacks:*"}, domain: orgABC, resource: ResourceClusters, action: ActionList, shouldAllow: false,
				}),
			)
		})
	})

	Context("Edge cases", func() {
		var env *testEnv

		BeforeEach(func() {
			env = newTestEnv(GinkgoT(), defaultTeams())
			env.policyMgr.AddGroupingPolicy(userAdmin, string(models.OrgAdminRole), orgABC)
		})

		It("should deny empty domain even for OrgAdmin", func() {
			ctx := ctxWithIdentity(jwtIdentity(userAdmin, orgABC))
			err := env.permService.Check(ctx, "", ResourceStacks, "", ActionList)
			Expect(err).ToNot(BeNil())
		})

		It("should allow OrgAdmin with empty resource via wildcard", func() {
			ctx := ctxWithIdentity(jwtIdentity(userAdmin, orgABC))
			err := env.permService.Check(ctx, orgABC, "", "", ActionList)
			Expect(err).To(BeNil())
		})
	})
})
