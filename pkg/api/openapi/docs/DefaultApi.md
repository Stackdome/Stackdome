# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ApiV1ApiTokensGet**](DefaultApi.md#ApiV1ApiTokensGet) | **Get** /api/v1/api-tokens | List all API tokens for the current user
[**ApiV1ApiTokensIdDelete**](DefaultApi.md#ApiV1ApiTokensIdDelete) | **Delete** /api/v1/api-tokens/{id} | Revoke an API token
[**ApiV1ApiTokensIdGet**](DefaultApi.md#ApiV1ApiTokensIdGet) | **Get** /api/v1/api-tokens/{id} | Get an API token by ID
[**ApiV1ApiTokensPost**](DefaultApi.md#ApiV1ApiTokensPost) | **Post** /api/v1/api-tokens | Create a new API token
[**ApiV1ApiTokensScopesGet**](DefaultApi.md#ApiV1ApiTokensScopesGet) | **Get** /api/v1/api-tokens/scopes | List all available API token scopes
[**ApiV1AuthGithubCallbackGet**](DefaultApi.md#ApiV1AuthGithubCallbackGet) | **Get** /api/v1/auth/github/callback | GitHub OAuth callback
[**ApiV1AuthGithubGet**](DefaultApi.md#ApiV1AuthGithubGet) | **Get** /api/v1/auth/github | Initiate GitHub OAuth flow
[**ApiV1AuthLoginPost**](DefaultApi.md#ApiV1AuthLoginPost) | **Post** /api/v1/auth/login | User login
[**ApiV1AuthRefreshPost**](DefaultApi.md#ApiV1AuthRefreshPost) | **Post** /api/v1/auth/refresh | Refresh JWT token
[**ApiV1ConfigGet**](DefaultApi.md#ApiV1ConfigGet) | **Get** /api/v1/config | Get public application configuration
[**ApiV1GitIntegrationsGithubManifestCallbackGet**](DefaultApi.md#ApiV1GitIntegrationsGithubManifestCallbackGet) | **Get** /api/v1/git-integrations/github/manifest/callback | GitHub App manifest redirect target (unauthenticated, state-validated)
[**ApiV1InvitesTokenInfoGet**](DefaultApi.md#ApiV1InvitesTokenInfoGet) | **Get** /api/v1/invites/{token}/info | Get public invite info (unauthenticated)
[**ApiV1OrganizationsIdGet**](DefaultApi.md#ApiV1OrganizationsIdGet) | **Get** /api/v1/organizations/{id} | Get an organization
[**ApiV1OrganizationsIdPut**](DefaultApi.md#ApiV1OrganizationsIdPut) | **Put** /api/v1/organizations/{id} | Update an organization
[**ApiV1OrganizationsOrgIdAdminsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdAdminsGet) | **Get** /api/v1/organizations/{org_id}/admins | List organization admins
[**ApiV1OrganizationsOrgIdAdminsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdAdminsPost) | **Post** /api/v1/organizations/{org_id}/admins | Promote a user to organization admin
[**ApiV1OrganizationsOrgIdAdminsUserIdDemotePost**](DefaultApi.md#ApiV1OrganizationsOrgIdAdminsUserIdDemotePost) | **Post** /api/v1/organizations/{org_id}/admins/{user_id}/demote | Demote an organization admin
[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet) | **Get** /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries | List all image registries for a cluster
[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete) | **Delete** /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries/{id} | Delete an image registry
[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet) | **Get** /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries/{id} | Get a specific image registry
[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost) | **Post** /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries | Create a new image registry
[**ApiV1OrganizationsOrgIdClustersGet**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersGet) | **Get** /api/v1/organizations/{org_id}/clusters | List all clusters for an organization
[**ApiV1OrganizationsOrgIdClustersIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersIdDelete) | **Delete** /api/v1/organizations/{org_id}/clusters/{id} | Delete a cluster
[**ApiV1OrganizationsOrgIdClustersIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersIdGet) | **Get** /api/v1/organizations/{org_id}/clusters/{id} | Get a specific cluster
[**ApiV1OrganizationsOrgIdClustersPost**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersPost) | **Post** /api/v1/organizations/{org_id}/clusters | Add a new cluster
[**ApiV1OrganizationsOrgIdGitIntegrationsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsGet) | **Get** /api/v1/organizations/{org_id}/git-integrations | List git integrations for the organization
[**ApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPost**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPost) | **Post** /api/v1/organizations/{org_id}/git-integrations/github/manifest | Start the GitHub App manifest flow for the organization
[**ApiV1OrganizationsOrgIdGitIntegrationsIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsIdDelete) | **Delete** /api/v1/organizations/{org_id}/git-integrations/{id} | Delete a git integration
[**ApiV1OrganizationsOrgIdGitIntegrationsIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsIdGet) | **Get** /api/v1/organizations/{org_id}/git-integrations/{id} | Get a git integration
[**ApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGet) | **Get** /api/v1/organizations/{org_id}/git-integrations/{id}/installations | List GitHub App installations
[**ApiV1OrganizationsOrgIdGitIntegrationsIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsIdPut) | **Put** /api/v1/organizations/{org_id}/git-integrations/{id} | Update a git integration (credential rotation)
[**ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGet) | **Get** /api/v1/organizations/{org_id}/git-integrations/{id}/repositories | List repositories visible to the GitHub App installation
[**ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGet) | **Get** /api/v1/organizations/{org_id}/git-integrations/{id}/repositories/{owner}/{repo}/branches | List repository branches through the GitHub App installation
[**ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGet**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGet) | **Get** /api/v1/organizations/{org_id}/git-integrations/{id}/repositories/{owner}/{repo} | Get repository details through the GitHub App installation
[**ApiV1OrganizationsOrgIdGitIntegrationsIdVerifyPost**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsIdVerifyPost) | **Post** /api/v1/organizations/{org_id}/git-integrations/{id}/verify | Verify a git integration against a repository
[**ApiV1OrganizationsOrgIdGitIntegrationsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdGitIntegrationsPost) | **Post** /api/v1/organizations/{org_id}/git-integrations | Create a git integration for the organization
[**ApiV1OrganizationsOrgIdImageRegistriesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdImageRegistriesGet) | **Get** /api/v1/organizations/{org_id}/image_registries | List all image registries for an organisation
[**ApiV1OrganizationsOrgIdInvitesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesGet) | **Get** /api/v1/organizations/{org_id}/invites | List invites for an organization
[**ApiV1OrganizationsOrgIdInvitesIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesIdDelete) | **Delete** /api/v1/organizations/{org_id}/invites/{id} | Revoke a pending invite
[**ApiV1OrganizationsOrgIdInvitesIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesIdGet) | **Get** /api/v1/organizations/{org_id}/invites/{id} | Get an invite by ID
[**ApiV1OrganizationsOrgIdInvitesIdResendPost**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesIdResendPost) | **Post** /api/v1/organizations/{org_id}/invites/{id}/resend | Re-queue invite email for delivery
[**ApiV1OrganizationsOrgIdInvitesPost**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesPost) | **Post** /api/v1/organizations/{org_id}/invites | Create an invite to the organization
[**ApiV1OrganizationsOrgIdObjectStoresGet**](DefaultApi.md#ApiV1OrganizationsOrgIdObjectStoresGet) | **Get** /api/v1/organizations/{org_id}/object-stores | List all object stores the user has access to across all projects
[**ApiV1OrganizationsOrgIdProjectsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsGet) | **Get** /api/v1/organizations/{org_id}/projects | List all projects in an organization
[**ApiV1OrganizationsOrgIdProjectsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsPost) | **Post** /api/v1/organizations/{org_id}/projects | Create a new project
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres | List all PostgresAddons for a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id}/actions/backup | Trigger an immediate backup
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id}/actions/fence | Fence or unfence the PostgreSQL cluster
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id}/actions/hibernate | Hibernate or wake the PostgreSQL cluster
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id}/backups | List backups for a PostgresAddon
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id}/credentials/{database} | Get JIT credentials for a PostgresAddon database
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id} | Delete a PostgresAddon
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id} | Get a specific PostgresAddon
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPut) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres/{id} | Update a PostgresAddon
[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/addons/postgres | Create a new PostgresAddon
[**ApiV1OrganizationsOrgIdProjectsProjectNameDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name} | Delete a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name} | Get a specific project by name
[**ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/members | List members of a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/members/{id} | Remove a member from a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPut) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/members/{id} | Update a project member&#39;s role
[**ApiV1OrganizationsOrgIdProjectsProjectNameMembersPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameMembersPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/members | Add a member to a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/object-stores | List all ObjectStores for a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/object-stores/{id} | Delete an ObjectStore
[**ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/object-stores/{id} | Get a specific ObjectStore
[**ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPut) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/object-stores/{id} | Update an ObjectStore
[**ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/object-stores | Add a new ObjectStore for backup storage
[**ApiV1OrganizationsOrgIdProjectsProjectNamePut**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNamePut) | **Put** /api/v1/organizations/{org_id}/projects/{project_name} | Update a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameSecretsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameSecretsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/secrets | List all secrets for a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/secrets/{id} | Delete a secret
[**ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/secrets/{id} | Get a specific secret
[**ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPut) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/secrets/{id} | Update a secret
[**ApiV1OrganizationsOrgIdProjectsProjectNameSecretsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameSecretsPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/secrets | Create a new secret
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks | List all stacks for a project
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/builds/{build_id} | Get a specific build under a stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/builds | List all builds under a stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/connections/{connection_id} | Delete stack connection
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPut) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/connections/{connection_id} | Update stack connection
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/connections | List stack connections
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/connections | Create stack connection
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id} | Delete a stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id} | Get a specific stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/logs | Get logs for a stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/metrics | Get metrics for a stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPut) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id} | Update a stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources | List all stack resources under a stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources/{resource_name}/actions/restart | Restart a stack resource
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources/{resource_name}/builds | List all builds for a stack resource
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources/{resource_name} | Get a specific stack resource by name
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources/{resource_name}/logs | Get logs for a stack resource
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources/{resource_name}/metrics | Get metrics for a stack resource
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/topology | Get stack topology
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/volumes | Create a volume and associate it with the stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/stacks | Create a new stack
[**ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/volumes/{id} | Delete a volume
[**ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/volumes/{id} | Get a specific volume
[**ApiV1OrganizationsOrgIdProjectsProjectNameVolumesPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameVolumesPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/volumes | Create a new volume
[**ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/workspace-users/current | Get the workspace user for the current user
[**ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdDelete) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/workspace-users/{id} | Delete a workspace user
[**ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGet) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/workspace-users/{id} | Get a workspace user by ID
[**ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPut) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/workspace-users/{id} | Update a workspace user
[**ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPost**](DefaultApi.md#ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPost) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/workspace-users | Create a new workspace user
[**ApiV1OrganizationsOrgIdRegistryCredentialsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdRegistryCredentialsGet) | **Get** /api/v1/organizations/{org_id}/registry-credentials | List registry credentials for the organization
[**ApiV1OrganizationsOrgIdRegistryCredentialsIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdRegistryCredentialsIdDelete) | **Delete** /api/v1/organizations/{org_id}/registry-credentials/{id} | Delete a registry credential
[**ApiV1OrganizationsOrgIdRegistryCredentialsIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdRegistryCredentialsIdGet) | **Get** /api/v1/organizations/{org_id}/registry-credentials/{id} | Get a registry credential
[**ApiV1OrganizationsOrgIdRegistryCredentialsIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdRegistryCredentialsIdPut) | **Put** /api/v1/organizations/{org_id}/registry-credentials/{id} | Update a registry credential (username/password rotation)
[**ApiV1OrganizationsOrgIdRegistryCredentialsIdVerifyPost**](DefaultApi.md#ApiV1OrganizationsOrgIdRegistryCredentialsIdVerifyPost) | **Post** /api/v1/organizations/{org_id}/registry-credentials/{id}/verify | Verify a registry credential against a repository
[**ApiV1OrganizationsOrgIdRegistryCredentialsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdRegistryCredentialsPost) | **Post** /api/v1/organizations/{org_id}/registry-credentials | Create a registry credential for the organization
[**ApiV1OrganizationsOrgIdSecretsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdSecretsGet) | **Get** /api/v1/organizations/{org_id}/secrets | List all secrets the user has access to across all projects
[**ApiV1OrganizationsOrgIdStacksGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksGet) | **Get** /api/v1/organizations/{org_id}/stacks | List all stacks the user has access to across all projects
[**ApiV1OrganizationsOrgIdUsersGet**](DefaultApi.md#ApiV1OrganizationsOrgIdUsersGet) | **Get** /api/v1/organizations/{org_id}/users | List all users in an organization
[**ApiV1ProjectRolesGet**](DefaultApi.md#ApiV1ProjectRolesGet) | **Get** /api/v1/project-roles | List available project membership roles
[**ApiV1UserSignupPost**](DefaultApi.md#ApiV1UserSignupPost) | **Post** /api/v1/user-signup | Create new user
[**ApiV1UsersCurrentGet**](DefaultApi.md#ApiV1UsersCurrentGet) | **Get** /api/v1/users/current | Get the current authenticated user
[**ApiV1UsersCurrentProjectsGet**](DefaultApi.md#ApiV1UsersCurrentProjectsGet) | **Get** /api/v1/users/current/projects | List projects for the current authenticated user
[**ApiV1UsersIdGet**](DefaultApi.md#ApiV1UsersIdGet) | **Get** /api/v1/users/{id} | Get a user
[**ApiV1WebhooksGithubPost**](DefaultApi.md#ApiV1WebhooksGithubPost) | **Post** /api/v1/webhooks/github | GitHub webhook receiver (unauthenticated, HMAC-verified)
[**ApplyStack**](DefaultApi.md#ApplyStack) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/apply | Apply a full stack document (declarative reconcile)
[**ApplyStackByName**](DefaultApi.md#ApplyStackByName) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/apply | Apply a full stack document by name (declarative upsert)
[**CreateStackResource**](DefaultApi.md#CreateStackResource) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources | Create a stack resource
[**DeleteStackResource**](DefaultApi.md#DeleteStackResource) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources/{resource_name} | Delete a stack resource
[**UpdateStackResource**](DefaultApi.md#UpdateStackResource) | **Put** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/resources/{resource_name} | Update a stack resource



## ApiV1ApiTokensGet

> APITokenList ApiV1ApiTokensGet(ctx).Execute()

List all API tokens for the current user

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ApiTokensGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ApiTokensGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ApiTokensGet`: APITokenList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ApiTokensGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ApiTokensGetRequest struct via the builder pattern


### Return type

[**APITokenList**](APITokenList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ApiTokensIdDelete

> ApiV1ApiTokensIdDelete(ctx, id).Execute()

Revoke an API token

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ApiTokensIdDelete(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ApiTokensIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ApiTokensIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ApiTokensIdGet

> APIToken ApiV1ApiTokensIdGet(ctx, id).Execute()

Get an API token by ID

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ApiTokensIdGet(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ApiTokensIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ApiTokensIdGet`: APIToken
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ApiTokensIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ApiTokensIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**APIToken**](APIToken.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ApiTokensPost

> APITokenCreateResponse ApiV1ApiTokensPost(ctx).APITokenCreateRequest(aPITokenCreateRequest).Execute()

Create a new API token

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    aPITokenCreateRequest := *openapiclient.NewAPITokenCreateRequest("Name_example", []string{"Scopes_example"}) // APITokenCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ApiTokensPost(context.Background()).APITokenCreateRequest(aPITokenCreateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ApiTokensPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ApiTokensPost`: APITokenCreateResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ApiTokensPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ApiTokensPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **aPITokenCreateRequest** | [**APITokenCreateRequest**](APITokenCreateRequest.md) |  | 

### Return type

[**APITokenCreateResponse**](APITokenCreateResponse.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ApiTokensScopesGet

> ScopeList ApiV1ApiTokensScopesGet(ctx).Execute()

List all available API token scopes



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ApiTokensScopesGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ApiTokensScopesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ApiTokensScopesGet`: ScopeList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ApiTokensScopesGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ApiTokensScopesGetRequest struct via the builder pattern


### Return type

[**ScopeList**](ScopeList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1AuthGithubCallbackGet

> LoginResponse ApiV1AuthGithubCallbackGet(ctx).Code(code).State(state).Execute()

GitHub OAuth callback



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    code := "code_example" // string | 
    state := "state_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1AuthGithubCallbackGet(context.Background()).Code(code).State(state).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1AuthGithubCallbackGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1AuthGithubCallbackGet`: LoginResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1AuthGithubCallbackGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1AuthGithubCallbackGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **code** | **string** |  | 
 **state** | **string** |  | 

### Return type

[**LoginResponse**](LoginResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1AuthGithubGet

> ApiV1AuthGithubGet(ctx).Execute()

Initiate GitHub OAuth flow



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1AuthGithubGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1AuthGithubGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1AuthGithubGetRequest struct via the builder pattern


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1AuthLoginPost

> LoginResponse ApiV1AuthLoginPost(ctx).LoginRequest(loginRequest).Execute()

User login



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    loginRequest := *openapiclient.NewLoginRequest("Email_example", "Password_example") // LoginRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1AuthLoginPost(context.Background()).LoginRequest(loginRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1AuthLoginPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1AuthLoginPost`: LoginResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1AuthLoginPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1AuthLoginPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **loginRequest** | [**LoginRequest**](LoginRequest.md) |  | 

### Return type

[**LoginResponse**](LoginResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1AuthRefreshPost

> RefreshTokenResponse ApiV1AuthRefreshPost(ctx).RefreshTokenRequest(refreshTokenRequest).Execute()

Refresh JWT token



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    refreshTokenRequest := *openapiclient.NewRefreshTokenRequest("RefreshToken_example") // RefreshTokenRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1AuthRefreshPost(context.Background()).RefreshTokenRequest(refreshTokenRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1AuthRefreshPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1AuthRefreshPost`: RefreshTokenResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1AuthRefreshPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1AuthRefreshPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **refreshTokenRequest** | [**RefreshTokenRequest**](RefreshTokenRequest.md) |  | 

### Return type

[**RefreshTokenResponse**](RefreshTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ConfigGet

> AppConfigResponse ApiV1ConfigGet(ctx).Execute()

Get public application configuration



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ConfigGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ConfigGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ConfigGet`: AppConfigResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ConfigGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ConfigGetRequest struct via the builder pattern


### Return type

[**AppConfigResponse**](AppConfigResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1GitIntegrationsGithubManifestCallbackGet

> ApiV1GitIntegrationsGithubManifestCallbackGet(ctx).Code(code).State(state).Execute()

GitHub App manifest redirect target (unauthenticated, state-validated)

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    code := "code_example" // string | 
    state := "state_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1GitIntegrationsGithubManifestCallbackGet(context.Background()).Code(code).State(state).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1GitIntegrationsGithubManifestCallbackGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1GitIntegrationsGithubManifestCallbackGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **code** | **string** |  | 
 **state** | **string** |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1InvitesTokenInfoGet

> OrgInviteInfo ApiV1InvitesTokenInfoGet(ctx, token).Execute()

Get public invite info (unauthenticated)

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    token := "token_example" // string | The invite token

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1InvitesTokenInfoGet(context.Background(), token).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1InvitesTokenInfoGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1InvitesTokenInfoGet`: OrgInviteInfo
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1InvitesTokenInfoGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**token** | **string** | The invite token | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1InvitesTokenInfoGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**OrgInviteInfo**](OrgInviteInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsIdGet

> Organisation ApiV1OrganizationsIdGet(ctx, id).Execute()

Get an organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsIdGet(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsIdGet`: Organisation
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Organisation**](Organisation.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsIdPut

> Organisation ApiV1OrganizationsIdPut(ctx, id).Organisation(organisation).Execute()

Update an organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The id of record
    organisation := *openapiclient.NewOrganisation() // Organisation | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsIdPut(context.Background(), id).Organisation(organisation).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsIdPut`: Organisation
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **organisation** | [**Organisation**](Organisation.md) |  | 

### Return type

[**Organisation**](Organisation.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdAdminsGet

> UserList ApiV1OrganizationsOrgIdAdminsGet(ctx, orgId).Execute()

List organization admins

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAdminsGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdAdminsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdAdminsGet`: UserList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdAdminsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdAdminsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**UserList**](UserList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdAdminsPost

> ApiV1OrganizationsOrgIdAdminsPost(ctx, orgId).PromoteAdminRequest(promoteAdminRequest).Execute()

Promote a user to organization admin

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    promoteAdminRequest := *openapiclient.NewPromoteAdminRequest("UserId_example") // PromoteAdminRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAdminsPost(context.Background(), orgId).PromoteAdminRequest(promoteAdminRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdAdminsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdAdminsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **promoteAdminRequest** | [**PromoteAdminRequest**](PromoteAdminRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdAdminsUserIdDemotePost

> ApiV1OrganizationsOrgIdAdminsUserIdDemotePost(ctx, orgId, userId).DemoteAdminRequest(demoteAdminRequest).Execute()

Demote an organization admin



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    userId := "userId_example" // string | The ID of the user
    demoteAdminRequest := *openapiclient.NewDemoteAdminRequest("ProjectName_example") // DemoteAdminRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAdminsUserIdDemotePost(context.Background(), orgId, userId).DemoteAdminRequest(demoteAdminRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdAdminsUserIdDemotePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**userId** | **string** | The ID of the user | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdAdminsUserIdDemotePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **demoteAdminRequest** | [**DemoteAdminRequest**](DemoteAdminRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet

> ClusterImageRegistryList ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet(ctx, orgId, clusterId).Execute()

List all image registries for a cluster

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    clusterId := "clusterId_example" // string | The ID of the cluster

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet(context.Background(), orgId, clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet`: ClusterImageRegistryList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**clusterId** | **string** | The ID of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ClusterImageRegistryList**](ClusterImageRegistryList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete

> ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete(ctx, orgId, clusterId, id).Execute()

Delete an image registry

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    clusterId := "clusterId_example" // string | The ID of the cluster
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete(context.Background(), orgId, clusterId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**clusterId** | **string** | The ID of the cluster | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet

> ClusterImageRegistry ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet(ctx, orgId, clusterId, id).Execute()

Get a specific image registry

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    clusterId := "clusterId_example" // string | The ID of the cluster
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet(context.Background(), orgId, clusterId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet`: ClusterImageRegistry
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**clusterId** | **string** | The ID of the cluster | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**ClusterImageRegistry**](ClusterImageRegistry.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost

> ClusterImageRegistry ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost(ctx, orgId, clusterId).ClusterImageRegistry(clusterImageRegistry).Execute()

Create a new image registry

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    clusterId := "clusterId_example" // string | The ID of the cluster
    clusterImageRegistry := *openapiclient.NewClusterImageRegistry("Name_example") // ClusterImageRegistry | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost(context.Background(), orgId, clusterId).ClusterImageRegistry(clusterImageRegistry).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost`: ClusterImageRegistry
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**clusterId** | **string** | The ID of the cluster | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **clusterImageRegistry** | [**ClusterImageRegistry**](ClusterImageRegistry.md) |  | 

### Return type

[**ClusterImageRegistry**](ClusterImageRegistry.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersGet

> ClusterList ApiV1OrganizationsOrgIdClustersGet(ctx, orgId).Execute()

List all clusters for an organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdClustersGet`: ClusterList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdClustersGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ClusterList**](ClusterList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersIdDelete

> ApiV1OrganizationsOrgIdClustersIdDelete(ctx, orgId, id).Execute()

Delete a cluster

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersIdDelete(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersIdGet

> Cluster ApiV1OrganizationsOrgIdClustersIdGet(ctx, orgId, id).Execute()

Get a specific cluster

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersIdGet(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdClustersIdGet`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdClustersIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Cluster**](Cluster.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersPost

> Cluster ApiV1OrganizationsOrgIdClustersPost(ctx, orgId).Cluster(cluster).Execute()

Add a new cluster

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    cluster := *openapiclient.NewCluster("Name_example", "ClusterUrl_example", "ClusterCaData_example", "ClusterSaToken_example") // Cluster | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersPost(context.Background(), orgId).Cluster(cluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdClustersPost`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdClustersPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **cluster** | [**Cluster**](Cluster.md) |  | 

### Return type

[**Cluster**](Cluster.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsGet

> GitIntegrationList ApiV1OrganizationsOrgIdGitIntegrationsGet(ctx, orgId).Execute()

List git integrations for the organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsGet`: GitIntegrationList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GitIntegrationList**](GitIntegrationList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPost

> GitHubAppManifestFlow ApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPost(ctx, orgId).Execute()

Start the GitHub App manifest flow for the organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPost(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPost`: GitHubAppManifestFlow
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsGithubManifestPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**GitHubAppManifestFlow**](GitHubAppManifestFlow.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsIdDelete

> ApiV1OrganizationsOrgIdGitIntegrationsIdDelete(ctx, orgId, id).Execute()

Delete a git integration

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdDelete(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsIdGet

> GitIntegration ApiV1OrganizationsOrgIdGitIntegrationsIdGet(ctx, orgId, id).Execute()

Get a git integration

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdGet(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsIdGet`: GitIntegration
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**GitIntegration**](GitIntegration.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGet

> GitInstallationList ApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGet(ctx, orgId, id).Refresh(refresh).Execute()

List GitHub App installations

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record
    refresh := true // bool | Re-list installations from GitHub before returning (covers missed webhooks) (optional) (default to false)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGet(context.Background(), orgId, id).Refresh(refresh).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGet`: GitInstallationList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsIdInstallationsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **refresh** | **bool** | Re-list installations from GitHub before returning (covers missed webhooks) | [default to false]

### Return type

[**GitInstallationList**](GitInstallationList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsIdPut

> GitIntegration ApiV1OrganizationsOrgIdGitIntegrationsIdPut(ctx, orgId, id).GitIntegration(gitIntegration).Execute()

Update a git integration (credential rotation)

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record
    gitIntegration := *openapiclient.NewGitIntegration("Host_example") // GitIntegration | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdPut(context.Background(), orgId, id).GitIntegration(gitIntegration).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsIdPut`: GitIntegration
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **gitIntegration** | [**GitIntegration**](GitIntegration.md) |  | 

### Return type

[**GitIntegration**](GitIntegration.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGet

> GitRepositoryPage ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGet(ctx, orgId, id).Page(page).InstallationId(installationId).Execute()

List repositories visible to the GitHub App installation

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record
    page := int32(56) // int32 |  (optional)
    installationId := "installationId_example" // string | Our GitInstallation id (UUID). When omitted, repositories are aggregated across every installation of the integration. (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGet(context.Background(), orgId, id).Page(page).InstallationId(installationId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGet`: GitRepositoryPage
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **page** | **int32** |  | 
 **installationId** | **string** | Our GitInstallation id (UUID). When omitted, repositories are aggregated across every installation of the integration. | 

### Return type

[**GitRepositoryPage**](GitRepositoryPage.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGet

> GitBranchList ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGet(ctx, orgId, id, owner, repo).Execute()

List repository branches through the GitHub App installation

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record
    owner := "owner_example" // string | 
    repo := "repo_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGet(context.Background(), orgId, id, owner, repo).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGet`: GitBranchList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 
**owner** | **string** |  | 
**repo** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoBranchesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**GitBranchList**](GitBranchList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGet

> GitRepository ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGet(ctx, orgId, id, owner, repo).Execute()

Get repository details through the GitHub App installation

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record
    owner := "owner_example" // string | 
    repo := "repo_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGet(context.Background(), orgId, id, owner, repo).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGet`: GitRepository
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 
**owner** | **string** |  | 
**repo** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsIdRepositoriesOwnerRepoGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**GitRepository**](GitRepository.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsIdVerifyPost

> ApiV1OrganizationsOrgIdGitIntegrationsIdVerifyPost(ctx, orgId, id).GitIntegrationVerifyRequest(gitIntegrationVerifyRequest).Execute()

Verify a git integration against a repository

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record
    gitIntegrationVerifyRequest := *openapiclient.NewGitIntegrationVerifyRequest("RepoUrl_example") // GitIntegrationVerifyRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdVerifyPost(context.Background(), orgId, id).GitIntegrationVerifyRequest(gitIntegrationVerifyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdVerifyPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsIdVerifyPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **gitIntegrationVerifyRequest** | [**GitIntegrationVerifyRequest**](GitIntegrationVerifyRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdGitIntegrationsPost

> GitIntegration ApiV1OrganizationsOrgIdGitIntegrationsPost(ctx, orgId).GitIntegration(gitIntegration).Execute()

Create a git integration for the organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    gitIntegration := *openapiclient.NewGitIntegration("Host_example") // GitIntegration | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsPost(context.Background(), orgId).GitIntegration(gitIntegration).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdGitIntegrationsPost`: GitIntegration
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdGitIntegrationsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **gitIntegration** | [**GitIntegration**](GitIntegration.md) |  | 

### Return type

[**GitIntegration**](GitIntegration.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdImageRegistriesGet

> ClusterImageRegistryList ApiV1OrganizationsOrgIdImageRegistriesGet(ctx, orgId).Execute()

List all image registries for an organisation

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdImageRegistriesGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdImageRegistriesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdImageRegistriesGet`: ClusterImageRegistryList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdImageRegistriesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdImageRegistriesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ClusterImageRegistryList**](ClusterImageRegistryList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdInvitesGet

> OrgInviteList ApiV1OrganizationsOrgIdInvitesGet(ctx, orgId).Status(status).Execute()

List invites for an organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    status := "status_example" // string | Filter invites by status (pending, accepted, revoked, expired) (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdInvitesGet(context.Background(), orgId).Status(status).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdInvitesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdInvitesGet`: OrgInviteList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdInvitesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdInvitesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **status** | **string** | Filter invites by status (pending, accepted, revoked, expired) | 

### Return type

[**OrgInviteList**](OrgInviteList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdInvitesIdDelete

> ApiV1OrganizationsOrgIdInvitesIdDelete(ctx, orgId, id).Execute()

Revoke a pending invite

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdDelete(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdInvitesIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdInvitesIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdInvitesIdGet

> OrgInvite ApiV1OrganizationsOrgIdInvitesIdGet(ctx, orgId, id).Execute()

Get an invite by ID

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdGet(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdInvitesIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdInvitesIdGet`: OrgInvite
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdInvitesIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdInvitesIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**OrgInvite**](OrgInvite.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdInvitesIdResendPost

> ApiV1OrganizationsOrgIdInvitesIdResendPost(ctx, orgId, id).Execute()

Re-queue invite email for delivery

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdResendPost(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdInvitesIdResendPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdInvitesIdResendPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdInvitesPost

> OrgInviteCreateResponse ApiV1OrganizationsOrgIdInvitesPost(ctx, orgId).OrgInviteCreateRequest(orgInviteCreateRequest).Execute()

Create an invite to the organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    orgInviteCreateRequest := *openapiclient.NewOrgInviteCreateRequest("Email_example", "ProjectName_example", "Role_example", int32(123)) // OrgInviteCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdInvitesPost(context.Background(), orgId).OrgInviteCreateRequest(orgInviteCreateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdInvitesPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdInvitesPost`: OrgInviteCreateResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdInvitesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdInvitesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **orgInviteCreateRequest** | [**OrgInviteCreateRequest**](OrgInviteCreateRequest.md) |  | 

### Return type

[**OrgInviteCreateResponse**](OrgInviteCreateResponse.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdObjectStoresGet

> ObjectStoreList ApiV1OrganizationsOrgIdObjectStoresGet(ctx, orgId).Execute()

List all object stores the user has access to across all projects



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdObjectStoresGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdObjectStoresGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdObjectStoresGet`: ObjectStoreList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdObjectStoresGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdObjectStoresGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ObjectStoreList**](ObjectStoreList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsGet

> ProjectList ApiV1OrganizationsOrgIdProjectsGet(ctx, orgId).Execute()

List all projects in an organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsGet`: ProjectList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ProjectList**](ProjectList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsPost

> Project ApiV1OrganizationsOrgIdProjectsPost(ctx, orgId).ProjectCreateRequest(projectCreateRequest).Execute()

Create a new project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectCreateRequest := *openapiclient.NewProjectCreateRequest("Name_example") // ProjectCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsPost(context.Background(), orgId).ProjectCreateRequest(projectCreateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsPost`: Project
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **projectCreateRequest** | [**ProjectCreateRequest**](ProjectCreateRequest.md) |  | 

### Return type

[**Project**](Project.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGet

> PostgresAddonList ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGet(ctx, orgId, projectName).Execute()

List all PostgresAddons for a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGet(context.Background(), orgId, projectName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGet`: PostgresAddonList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**PostgresAddonList**](PostgresAddonList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost

> ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost202Response ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost(ctx, orgId, projectName, id).ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest(apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest).Execute()

Trigger an immediate backup

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest := *openapiclient.NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest() // ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost(context.Background(), orgId, projectName, id).ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest(apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost`: ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost202Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest** | [**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest**](ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPostRequest.md) |  | 

### Return type

[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost202Response**](ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsBackupPost202Response.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost

> ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost200Response ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost(ctx, orgId, projectName, id).ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest(apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest).Execute()

Fence or unfence the PostgreSQL cluster

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest := *openapiclient.NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest(false) // ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost(context.Background(), orgId, projectName, id).ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest(apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost`: ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost200Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest** | [**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest**](ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest.md) |  | 

### Return type

[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost200Response**](ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePost200Response.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost

> ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost200Response ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost(ctx, orgId, projectName, id).ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest(apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest).Execute()

Hibernate or wake the PostgreSQL cluster

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest := *openapiclient.NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest(false) // ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost(context.Background(), orgId, projectName, id).ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest(apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost`: ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost200Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **apiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest** | [**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest**](ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePostRequest.md) |  | 

### Return type

[**ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost200Response**](ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsHibernatePost200Response.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet

> PostgresBackupList ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet(ctx, orgId, projectName, id).Limit(limit).Offset(offset).Execute()

List backups for a PostgresAddon

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    limit := int32(56) // int32 |  (optional) (default to 20)
    offset := int32(56) // int32 |  (optional) (default to 0)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet(context.Background(), orgId, projectName, id).Limit(limit).Offset(offset).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet`: PostgresBackupList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdBackupsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **limit** | **int32** |  | [default to 20]
 **offset** | **int32** |  | [default to 0]

### Return type

[**PostgresBackupList**](PostgresBackupList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet

> PostgresCredentials ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet(ctx, orgId, projectName, id, database).Superuser(superuser).Execute()

Get JIT credentials for a PostgresAddon database

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    database := "database_example" // string | The name of the database
    superuser := true // bool | Whether to return superuser credentials (optional) (default to false)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet(context.Background(), orgId, projectName, id, database).Superuser(superuser).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet`: PostgresCredentials
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**database** | **string** | The name of the database | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdCredentialsDatabaseGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **superuser** | **bool** | Whether to return superuser credentials | [default to false]

### Return type

[**PostgresCredentials**](PostgresCredentials.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDelete

> PostgresAddon ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDelete(ctx, orgId, projectName, id).Execute()

Delete a PostgresAddon

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDelete(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDelete`: PostgresAddon
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**PostgresAddon**](PostgresAddon.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet

> PostgresAddon ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet(ctx, orgId, projectName, id).Execute()

Get a specific PostgresAddon

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet`: PostgresAddon
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**PostgresAddon**](PostgresAddon.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPut

> PostgresAddon ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPut(ctx, orgId, projectName, id).PostgresAddon(postgresAddon).Execute()

Update a PostgresAddon

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    postgresAddon := *openapiclient.NewPostgresAddon("Name_example", *openapiclient.NewPostgresAddonSpec(*openapiclient.NewPostgresVersion(int32(123)), *openapiclient.NewPostgresInstances(int32(123)), *openapiclient.NewPostgresStorage("Size_example"))) // PostgresAddon | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPut(context.Background(), orgId, projectName, id).PostgresAddon(postgresAddon).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPut`: PostgresAddon
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **postgresAddon** | [**PostgresAddon**](PostgresAddon.md) |  | 

### Return type

[**PostgresAddon**](PostgresAddon.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPost

> PostgresAddon ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPost(ctx, orgId, projectName).PostgresAddon(postgresAddon).Execute()

Create a new PostgresAddon



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    postgresAddon := *openapiclient.NewPostgresAddon("Name_example", *openapiclient.NewPostgresAddonSpec(*openapiclient.NewPostgresVersion(int32(123)), *openapiclient.NewPostgresInstances(int32(123)), *openapiclient.NewPostgresStorage("Size_example"))) // PostgresAddon | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPost(context.Background(), orgId, projectName).PostgresAddon(postgresAddon).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPost`: PostgresAddon
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **postgresAddon** | [**PostgresAddon**](PostgresAddon.md) |  | 

### Return type

[**PostgresAddon**](PostgresAddon.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameDelete

> ApiV1OrganizationsOrgIdProjectsProjectNameDelete(ctx, orgId, projectName).Execute()

Delete a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameDelete(context.Background(), orgId, projectName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameGet

> Project ApiV1OrganizationsOrgIdProjectsProjectNameGet(ctx, orgId, projectName).Execute()

Get a specific project by name

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameGet(context.Background(), orgId, projectName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameGet`: Project
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Project**](Project.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet

> ProjectMembershipList ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet(ctx, orgId, projectName).Execute()

List members of a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet(context.Background(), orgId, projectName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet`: ProjectMembershipList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameMembersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ProjectMembershipList**](ProjectMembershipList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdDelete

> ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdDelete(ctx, orgId, projectName, id).Execute()

Remove a member from a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdDelete(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameMembersIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPut

> ProjectMembership ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPut(ctx, orgId, projectName, id).UpdateProjectMemberRoleRequest(updateProjectMemberRoleRequest).Execute()

Update a project member's role

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    updateProjectMemberRoleRequest := *openapiclient.NewUpdateProjectMemberRoleRequest(openapiclient.ProjectRole("Developer")) // UpdateProjectMemberRoleRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPut(context.Background(), orgId, projectName, id).UpdateProjectMemberRoleRequest(updateProjectMemberRoleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPut`: ProjectMembership
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameMembersIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **updateProjectMemberRoleRequest** | [**UpdateProjectMemberRoleRequest**](UpdateProjectMemberRoleRequest.md) |  | 

### Return type

[**ProjectMembership**](ProjectMembership.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameMembersPost

> ProjectMembership ApiV1OrganizationsOrgIdProjectsProjectNameMembersPost(ctx, orgId, projectName).AddProjectMemberRequest(addProjectMemberRequest).Execute()

Add a member to a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    addProjectMemberRequest := *openapiclient.NewAddProjectMemberRequest("UserId_example", openapiclient.ProjectRole("Developer")) // AddProjectMemberRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersPost(context.Background(), orgId, projectName).AddProjectMemberRequest(addProjectMemberRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameMembersPost`: ProjectMembership
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameMembersPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **addProjectMemberRequest** | [**AddProjectMemberRequest**](AddProjectMemberRequest.md) |  | 

### Return type

[**ProjectMembership**](ProjectMembership.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGet

> ObjectStoreList ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGet(ctx, orgId, projectName).Execute()

List all ObjectStores for a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGet(context.Background(), orgId, projectName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGet`: ObjectStoreList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ObjectStoreList**](ObjectStoreList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdDelete

> ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdDelete(ctx, orgId, projectName, id).Execute()

Delete an ObjectStore

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdDelete(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGet

> ObjectStore ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGet(ctx, orgId, projectName, id).Execute()

Get a specific ObjectStore

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGet`: ObjectStore
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**ObjectStore**](ObjectStore.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPut

> ObjectStore ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPut(ctx, orgId, projectName, id).ObjectStore(objectStore).Execute()

Update an ObjectStore

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    objectStore := *openapiclient.NewObjectStore("Name_example", *openapiclient.NewObjectStoreSpec(*openapiclient.NewObjectStoreConfiguration(), "DestinationPath_example")) // ObjectStore | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPut(context.Background(), orgId, projectName, id).ObjectStore(objectStore).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPut`: ObjectStore
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **objectStore** | [**ObjectStore**](ObjectStore.md) |  | 

### Return type

[**ObjectStore**](ObjectStore.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPost

> ObjectStore ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPost(ctx, orgId, projectName).ObjectStore(objectStore).Execute()

Add a new ObjectStore for backup storage



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    objectStore := *openapiclient.NewObjectStore("Name_example", *openapiclient.NewObjectStoreSpec(*openapiclient.NewObjectStoreConfiguration(), "DestinationPath_example")) // ObjectStore | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPost(context.Background(), orgId, projectName).ObjectStore(objectStore).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPost`: ObjectStore
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameObjectStoresPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **objectStore** | [**ObjectStore**](ObjectStore.md) |  | 

### Return type

[**ObjectStore**](ObjectStore.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNamePut

> Project ApiV1OrganizationsOrgIdProjectsProjectNamePut(ctx, orgId, projectName).ProjectUpdateRequest(projectUpdateRequest).Execute()

Update a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    projectUpdateRequest := *openapiclient.NewProjectUpdateRequest("Name_example") // ProjectUpdateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNamePut(context.Background(), orgId, projectName).ProjectUpdateRequest(projectUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNamePut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNamePut`: Project
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNamePut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNamePutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **projectUpdateRequest** | [**ProjectUpdateRequest**](ProjectUpdateRequest.md) |  | 

### Return type

[**Project**](Project.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameSecretsGet

> SecretList ApiV1OrganizationsOrgIdProjectsProjectNameSecretsGet(ctx, orgId, projectName).Name(name).Execute()

List all secrets for a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    name := "name_example" // string | Filter by secret name (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsGet(context.Background(), orgId, projectName).Name(name).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameSecretsGet`: SecretList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameSecretsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **name** | **string** | Filter by secret name | 

### Return type

[**SecretList**](SecretList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdDelete

> ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdDelete(ctx, orgId, projectName, id).Execute()

Delete a secret

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdDelete(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGet

> Secret ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGet(ctx, orgId, projectName, id).Execute()

Get a specific secret

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGet`: Secret
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**Secret**](Secret.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPut

> Secret ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPut(ctx, orgId, projectName, id).Secret(secret).Execute()

Update a secret

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    secret := *openapiclient.NewSecret("Name_example", openapiclient.SecretType("Generic"), []openapiclient.SecretData{*openapiclient.NewSecretData("Key_example", "Value_example")}) // Secret | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPut(context.Background(), orgId, projectName, id).Secret(secret).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPut`: Secret
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameSecretsIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **secret** | [**Secret**](Secret.md) |  | 

### Return type

[**Secret**](Secret.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameSecretsPost

> Secret ApiV1OrganizationsOrgIdProjectsProjectNameSecretsPost(ctx, orgId, projectName).Secret(secret).Execute()

Create a new secret

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    secret := *openapiclient.NewSecret("Name_example", openapiclient.SecretType("Generic"), []openapiclient.SecretData{*openapiclient.NewSecretData("Key_example", "Value_example")}) // Secret | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsPost(context.Background(), orgId, projectName).Secret(secret).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameSecretsPost`: Secret
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameSecretsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameSecretsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **secret** | [**Secret**](Secret.md) |  | 

### Return type

[**Secret**](Secret.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksGet

> StackList ApiV1OrganizationsOrgIdProjectsProjectNameStacksGet(ctx, orgId, projectName).Limit(limit).Offset(offset).Execute()

List all stacks for a project

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    limit := int32(56) // int32 |  (optional) (default to 20)
    offset := int32(56) // int32 |  (optional) (default to 0)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksGet(context.Background(), orgId, projectName).Limit(limit).Offset(offset).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksGet`: StackList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **limit** | **int32** |  | [default to 20]
 **offset** | **int32** |  | [default to 0]

### Return type

[**StackList**](StackList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGet

> ImageBuild ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGet(ctx, orgId, projectName, id, buildId).Execute()

Get a specific build under a stack

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    buildId := "buildId_example" // string | The ID of the build

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGet(context.Background(), orgId, projectName, id, buildId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGet`: ImageBuild
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**buildId** | **string** | The ID of the build | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsBuildIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**ImageBuild**](ImageBuild.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet

> ImageBuildList ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet(ctx, orgId, projectName, id).Execute()

List all builds under a stack

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet`: ImageBuildList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdBuildsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**ImageBuildList**](ImageBuildList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdDelete

> ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdDelete(ctx, orgId, projectName, id, connectionId).Execute()

Delete stack connection

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    connectionId := "connectionId_example" // string | The ID of the stack connection

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdDelete(context.Background(), orgId, projectName, id, connectionId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**connectionId** | **string** | The ID of the stack connection | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPut

> StackConnection ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPut(ctx, orgId, projectName, id, connectionId).StackConnection(stackConnection).Execute()

Update stack connection

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    connectionId := "connectionId_example" // string | The ID of the stack connection
    stackConnection := *openapiclient.NewStackConnection("Kind_example", *openapiclient.NewTopologyNodeRef("Type_example"), *openapiclient.NewTopologyNodeRef("Type_example")) // StackConnection | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPut(context.Background(), orgId, projectName, id, connectionId).StackConnection(stackConnection).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPut`: StackConnection
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**connectionId** | **string** | The ID of the stack connection | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsConnectionIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **stackConnection** | [**StackConnection**](StackConnection.md) |  | 

### Return type

[**StackConnection**](StackConnection.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGet

> StackConnectionList ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGet(ctx, orgId, projectName, id).Execute()

List stack connections

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGet`: StackConnectionList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**StackConnectionList**](StackConnectionList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPost

> StackConnection ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPost(ctx, orgId, projectName, id).StackConnection(stackConnection).Execute()

Create stack connection

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    stackConnection := *openapiclient.NewStackConnection("Kind_example", *openapiclient.NewTopologyNodeRef("Type_example"), *openapiclient.NewTopologyNodeRef("Type_example")) // StackConnection | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPost(context.Background(), orgId, projectName, id).StackConnection(stackConnection).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPost`: StackConnection
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdConnectionsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **stackConnection** | [**StackConnection**](StackConnection.md) |  | 

### Return type

[**StackConnection**](StackConnection.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDelete

> Stack ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDelete(ctx, orgId, projectName, id).Execute()

Delete a stack

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDelete(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDelete`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**Stack**](Stack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet

> Stack ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet(ctx, orgId, projectName, id).Execute()

Get a specific stack

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**Stack**](Stack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGet

> *os.File ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGet(ctx, orgId, projectName, id).Follow(follow).Tail(tail).Since(since).Execute()

Get logs for a stack

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    follow := true // bool |  (optional) (default to false)
    tail := int32(56) // int32 |  (optional) (default to 100)
    since := "since_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGet(context.Background(), orgId, projectName, id).Follow(follow).Tail(tail).Since(since).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGet`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdLogsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **follow** | **bool** |  | [default to false]
 **tail** | **int32** |  | [default to 100]
 **since** | **string** |  | 

### Return type

[***os.File**](*os.File.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/event-stream, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGet

> *os.File ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGet(ctx, orgId, projectName, id).Stream(stream).Execute()

Get metrics for a stack



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    stream := true // bool |  (optional) (default to false)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGet(context.Background(), orgId, projectName, id).Stream(stream).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGet`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdMetricsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **stream** | **bool** |  | [default to false]

### Return type

[***os.File**](*os.File.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/event-stream, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPut

> Stack ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPut(ctx, orgId, projectName, id).Stack(stack).Execute()

Update a stack



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    stack := *openapiclient.NewStack("Name_example", *openapiclient.NewStackSpec()) // Stack | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPut(context.Background(), orgId, projectName, id).Stack(stack).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPut`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **stack** | [**Stack**](Stack.md) |  | 

### Return type

[**Stack**](Stack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGet

> StackResourceList ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGet(ctx, orgId, projectName, id).Execute()

List all stack resources under a stack

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGet`: StackResourceList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**StackResourceList**](StackResourceList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPost

> StackResource ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPost(ctx, orgId, projectName, id, resourceName).Execute()

Restart a stack resource



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPost(context.Background(), orgId, projectName, id, resourceName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPost`: StackResource
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameActionsRestartPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**StackResource**](StackResource.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet

> ImageBuildList ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet(ctx, orgId, projectName, id, resourceName).Execute()

List all builds for a stack resource

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet(context.Background(), orgId, projectName, id, resourceName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet`: ImageBuildList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameBuildsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**ImageBuildList**](ImageBuildList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGet

> StackResource ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGet(ctx, orgId, projectName, id, resourceName).Execute()

Get a specific stack resource by name

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGet(context.Background(), orgId, projectName, id, resourceName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGet`: StackResource
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**StackResource**](StackResource.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGet

> *os.File ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGet(ctx, orgId, projectName, id, resourceName).Follow(follow).Tail(tail).Since(since).Execute()

Get logs for a stack resource

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource
    follow := true // bool |  (optional) (default to false)
    tail := int32(56) // int32 |  (optional) (default to 100)
    since := "since_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGet(context.Background(), orgId, projectName, id, resourceName).Follow(follow).Tail(tail).Since(since).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGet`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameLogsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **follow** | **bool** |  | [default to false]
 **tail** | **int32** |  | [default to 100]
 **since** | **string** |  | 

### Return type

[***os.File**](*os.File.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/event-stream, application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGet

> ResourceMetrics ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGet(ctx, orgId, projectName, id, resourceName).Stream(stream).Execute()

Get metrics for a stack resource



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource
    stream := true // bool | Whether to stream metrics via Server-Sent Events (SSE) (optional) (default to false)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGet(context.Background(), orgId, projectName, id, resourceName).Stream(stream).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGet`: ResourceMetrics
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdResourcesResourceNameMetricsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **stream** | **bool** | Whether to stream metrics via Server-Sent Events (SSE) | [default to false]

### Return type

[**ResourceMetrics**](ResourceMetrics.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json, text/event-stream

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGet

> StackTopology ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGet(ctx, orgId, projectName, id).Execute()

Get stack topology

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGet`: StackTopology
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdTopologyGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**StackTopology**](StackTopology.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost

> Volume ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost(ctx, orgId, projectName, id).Volume(volume).Execute()

Create a volume and associate it with the stack

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    volume := *openapiclient.NewVolume("Name_example", *openapiclient.NewVolumeSpec("Size_example", false, openapiclient.VolumeAccessMode("ReadWriteOnce"))) // Volume | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost(context.Background(), orgId, projectName, id).Volume(volume).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost`: Volume
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **volume** | [**Volume**](Volume.md) |  | 

### Return type

[**Volume**](Volume.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost

> Stack ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost(ctx, orgId, projectName).Stack(stack).Execute()

Create a new stack



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    stack := *openapiclient.NewStack("Name_example", *openapiclient.NewStackSpec()) // Stack | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost(context.Background(), orgId, projectName).Stack(stack).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameStacksPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **stack** | [**Stack**](Stack.md) |  | 

### Return type

[**Stack**](Stack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdDelete

> ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdDelete(ctx, orgId, projectName, id).Execute()

Delete a volume

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdDelete(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet

> Volume ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet(ctx, orgId, projectName, id).Execute()

Get a specific volume

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet`: Volume
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**Volume**](Volume.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameVolumesPost

> Volume ApiV1OrganizationsOrgIdProjectsProjectNameVolumesPost(ctx, orgId, projectName).Volume(volume).Execute()

Create a new volume

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    volume := *openapiclient.NewVolume("Name_example", *openapiclient.NewVolumeSpec("Size_example", false, openapiclient.VolumeAccessMode("ReadWriteOnce"))) // Volume | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameVolumesPost(context.Background(), orgId, projectName).Volume(volume).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameVolumesPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameVolumesPost`: Volume
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameVolumesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameVolumesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **volume** | [**Volume**](Volume.md) |  | 

### Return type

[**Volume**](Volume.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGet

> WorkspaceUser ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGet(ctx, orgId, projectName).Execute()

Get the workspace user for the current user

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGet(context.Background(), orgId, projectName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGet`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersCurrentGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**WorkspaceUser**](WorkspaceUser.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdDelete

> ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdDelete(ctx, orgId, projectName, id).Execute()

Delete a workspace user

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdDelete(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGet

> WorkspaceUser ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGet(ctx, orgId, projectName, id).Execute()

Get a workspace user by ID

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGet(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGet`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**WorkspaceUser**](WorkspaceUser.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPut

> WorkspaceUser ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPut(ctx, orgId, projectName, id).WorkspaceUser(workspaceUser).Execute()

Update a workspace user

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    workspaceUser := *openapiclient.NewWorkspaceUser([]string{"Workspaces_example"}) // WorkspaceUser | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPut(context.Background(), orgId, projectName, id).WorkspaceUser(workspaceUser).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPut`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **workspaceUser** | [**WorkspaceUser**](WorkspaceUser.md) |  | 

### Return type

[**WorkspaceUser**](WorkspaceUser.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPost

> WorkspaceUser ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPost(ctx, orgId, projectName).WorkspaceUser(workspaceUser).Execute()

Create a new workspace user

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    workspaceUser := *openapiclient.NewWorkspaceUser([]string{"Workspaces_example"}) // WorkspaceUser | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPost(context.Background(), orgId, projectName).WorkspaceUser(workspaceUser).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPost`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdProjectsProjectNameWorkspaceUsersPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **workspaceUser** | [**WorkspaceUser**](WorkspaceUser.md) |  | 

### Return type

[**WorkspaceUser**](WorkspaceUser.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdRegistryCredentialsGet

> RegistryCredentialList ApiV1OrganizationsOrgIdRegistryCredentialsGet(ctx, orgId).Execute()

List registry credentials for the organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdRegistryCredentialsGet`: RegistryCredentialList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRegistryCredentialsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**RegistryCredentialList**](RegistryCredentialList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdRegistryCredentialsIdDelete

> RegistryCredentialDeleteResponse ApiV1OrganizationsOrgIdRegistryCredentialsIdDelete(ctx, orgId, id).Execute()

Delete a registry credential



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdDelete(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdRegistryCredentialsIdDelete`: RegistryCredentialDeleteResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRegistryCredentialsIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RegistryCredentialDeleteResponse**](RegistryCredentialDeleteResponse.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdRegistryCredentialsIdGet

> RegistryCredential ApiV1OrganizationsOrgIdRegistryCredentialsIdGet(ctx, orgId, id).Execute()

Get a registry credential

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdGet(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdRegistryCredentialsIdGet`: RegistryCredential
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRegistryCredentialsIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RegistryCredential**](RegistryCredential.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdRegistryCredentialsIdPut

> RegistryCredential ApiV1OrganizationsOrgIdRegistryCredentialsIdPut(ctx, orgId, id).RegistryCredential(registryCredential).Execute()

Update a registry credential (username/password rotation)

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record
    registryCredential := *openapiclient.NewRegistryCredential("Host_example", "Username_example") // RegistryCredential | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdPut(context.Background(), orgId, id).RegistryCredential(registryCredential).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdRegistryCredentialsIdPut`: RegistryCredential
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRegistryCredentialsIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **registryCredential** | [**RegistryCredential**](RegistryCredential.md) |  | 

### Return type

[**RegistryCredential**](RegistryCredential.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdRegistryCredentialsIdVerifyPost

> ApiV1OrganizationsOrgIdRegistryCredentialsIdVerifyPost(ctx, orgId, id).RegistryCredentialVerifyRequest(registryCredentialVerifyRequest).Execute()

Verify a registry credential against a repository

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    id := "id_example" // string | The id of record
    registryCredentialVerifyRequest := *openapiclient.NewRegistryCredentialVerifyRequest("Repository_example") // RegistryCredentialVerifyRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdVerifyPost(context.Background(), orgId, id).RegistryCredentialVerifyRequest(registryCredentialVerifyRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsIdVerifyPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRegistryCredentialsIdVerifyPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **registryCredentialVerifyRequest** | [**RegistryCredentialVerifyRequest**](RegistryCredentialVerifyRequest.md) |  | 

### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdRegistryCredentialsPost

> RegistryCredential ApiV1OrganizationsOrgIdRegistryCredentialsPost(ctx, orgId).RegistryCredential(registryCredential).Execute()

Create a registry credential for the organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    registryCredential := *openapiclient.NewRegistryCredential("Host_example", "Username_example") // RegistryCredential | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsPost(context.Background(), orgId).RegistryCredential(registryCredential).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdRegistryCredentialsPost`: RegistryCredential
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdRegistryCredentialsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRegistryCredentialsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **registryCredential** | [**RegistryCredential**](RegistryCredential.md) |  | 

### Return type

[**RegistryCredential**](RegistryCredential.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdSecretsGet

> SecretList ApiV1OrganizationsOrgIdSecretsGet(ctx, orgId).Name(name).Execute()

List all secrets the user has access to across all projects



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    name := "name_example" // string | Filter by secret name (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdSecretsGet(context.Background(), orgId).Name(name).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdSecretsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdSecretsGet`: SecretList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdSecretsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdSecretsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **name** | **string** | Filter by secret name | 

### Return type

[**SecretList**](SecretList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdStacksGet

> StackList ApiV1OrganizationsOrgIdStacksGet(ctx, orgId).Execute()

List all stacks the user has access to across all projects



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksGet`: StackList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**StackList**](StackList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdUsersGet

> UserList ApiV1OrganizationsOrgIdUsersGet(ctx, orgId).Page(page).PageSize(pageSize).Execute()

List all users in an organization

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    page := int32(56) // int32 | Page number (optional) (default to 1)
    pageSize := int32(56) // int32 | Number of items per page (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdUsersGet(context.Background(), orgId).Page(page).PageSize(pageSize).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdUsersGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdUsersGet`: UserList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdUsersGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdUsersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **page** | **int32** | Page number | [default to 1]
 **pageSize** | **int32** | Number of items per page | [default to 20]

### Return type

[**UserList**](UserList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ProjectRolesGet

> ProjectRoleList ApiV1ProjectRolesGet(ctx).Execute()

List available project membership roles



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ProjectRolesGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ProjectRolesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ProjectRolesGet`: ProjectRoleList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ProjectRolesGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ProjectRolesGetRequest struct via the builder pattern


### Return type

[**ProjectRoleList**](ProjectRoleList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1UserSignupPost

> UserSignupResponse ApiV1UserSignupPost(ctx).UserSignupRequest(userSignupRequest).Execute()

Create new user



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    userSignupRequest := *openapiclient.NewUserSignupRequest("Name_example", "Email_example", "Password_example") // UserSignupRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1UserSignupPost(context.Background()).UserSignupRequest(userSignupRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1UserSignupPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1UserSignupPost`: UserSignupResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1UserSignupPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1UserSignupPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userSignupRequest** | [**UserSignupRequest**](UserSignupRequest.md) |  | 

### Return type

[**UserSignupResponse**](UserSignupResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1UsersCurrentGet

> User ApiV1UsersCurrentGet(ctx).Execute()

Get the current authenticated user

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1UsersCurrentGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1UsersCurrentGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1UsersCurrentGet`: User
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1UsersCurrentGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1UsersCurrentGetRequest struct via the builder pattern


### Return type

[**User**](User.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1UsersCurrentProjectsGet

> ProjectList ApiV1UsersCurrentProjectsGet(ctx).Execute()

List projects for the current authenticated user

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1UsersCurrentProjectsGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1UsersCurrentProjectsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1UsersCurrentProjectsGet`: ProjectList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1UsersCurrentProjectsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1UsersCurrentProjectsGetRequest struct via the builder pattern


### Return type

[**ProjectList**](ProjectList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1UsersIdGet

> User ApiV1UsersIdGet(ctx, id).Execute()

Get a user



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1UsersIdGet(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1UsersIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1UsersIdGet`: User
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1UsersIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1UsersIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**User**](User.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1WebhooksGithubPost

> ApiV1WebhooksGithubPost(ctx).RequestBody(requestBody).Execute()

GitHub webhook receiver (unauthenticated, HMAC-verified)

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    requestBody := map[string]interface{}{"key": interface{}(123)} // map[string]interface{} | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1WebhooksGithubPost(context.Background()).RequestBody(requestBody).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WebhooksGithubPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WebhooksGithubPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **requestBody** | **map[string]interface{}** |  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApplyStack

> Stack ApplyStack(ctx, orgId, projectName, id).Stack(stack).Execute()

Apply a full stack document (declarative reconcile)



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    stack := *openapiclient.NewStack("Name_example", *openapiclient.NewStackSpec()) // Stack | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApplyStack(context.Background(), orgId, projectName, id).Stack(stack).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApplyStack``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApplyStack`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApplyStack`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApplyStackRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **stack** | [**Stack**](Stack.md) |  | 

### Return type

[**Stack**](Stack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApplyStackByName

> Stack ApplyStackByName(ctx, orgId, projectName).Stack(stack).Execute()

Apply a full stack document by name (declarative upsert)



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    stack := *openapiclient.NewStack("Name_example", *openapiclient.NewStackSpec()) // Stack | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApplyStackByName(context.Background(), orgId, projectName).Stack(stack).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApplyStackByName``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApplyStackByName`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApplyStackByName`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiApplyStackByNameRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **stack** | [**Stack**](Stack.md) |  | 

### Return type

[**Stack**](Stack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## CreateStackResource

> StackResource CreateStackResource(ctx, orgId, projectName, id).StackResource(stackResource).Execute()

Create a stack resource

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    stackResource := *openapiclient.NewStackResource("Name_example") // StackResource | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.CreateStackResource(context.Background(), orgId, projectName, id).StackResource(stackResource).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.CreateStackResource``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateStackResource`: StackResource
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.CreateStackResource`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreateStackResourceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **stackResource** | [**StackResource**](StackResource.md) |  | 

### Return type

[**StackResource**](StackResource.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeleteStackResource

> DeleteStackResource(ctx, orgId, projectName, id, resourceName).Execute()

Delete a stack resource

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.DeleteStackResource(context.Background(), orgId, projectName, id, resourceName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.DeleteStackResource``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeleteStackResourceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

 (empty response body)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdateStackResource

> StackResource UpdateStackResource(ctx, orgId, projectName, id, resourceName).StackResource(stackResource).Execute()

Update a stack resource

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    orgId := "orgId_example" // string | The ID of the organization
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource
    stackResource := *openapiclient.NewStackResource("Name_example") // StackResource | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.UpdateStackResource(context.Background(), orgId, projectName, id, resourceName).StackResource(stackResource).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.UpdateStackResource``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdateStackResource`: StackResource
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.UpdateStackResource`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiUpdateStackResourceRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **stackResource** | [**StackResource**](StackResource.md) |  | 

### Return type

[**StackResource**](StackResource.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

