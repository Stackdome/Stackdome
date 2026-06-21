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
[**ApiV1OrganizationsOrgIdInvitesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesGet) | **Get** /api/v1/organizations/{org_id}/invites | List invites for an organization
[**ApiV1OrganizationsOrgIdInvitesIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesIdDelete) | **Delete** /api/v1/organizations/{org_id}/invites/{id} | Revoke a pending invite
[**ApiV1OrganizationsOrgIdInvitesIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesIdGet) | **Get** /api/v1/organizations/{org_id}/invites/{id} | Get an invite by ID
[**ApiV1OrganizationsOrgIdInvitesIdResendPost**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesIdResendPost) | **Post** /api/v1/organizations/{org_id}/invites/{id}/resend | Re-queue invite email for delivery
[**ApiV1OrganizationsOrgIdInvitesPost**](DefaultApi.md#ApiV1OrganizationsOrgIdInvitesPost) | **Post** /api/v1/organizations/{org_id}/invites | Create an invite to the organization
[**ApiV1OrganizationsOrgIdObjectStoresGet**](DefaultApi.md#ApiV1OrganizationsOrgIdObjectStoresGet) | **Get** /api/v1/organizations/{org_id}/object-stores | List all object stores the user has access to across all teams
[**ApiV1OrganizationsOrgIdSecretsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdSecretsGet) | **Get** /api/v1/organizations/{org_id}/secrets | List all secrets the user has access to across all teams
[**ApiV1OrganizationsOrgIdStacksGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksGet) | **Get** /api/v1/organizations/{org_id}/stacks | List all stacks the user has access to across all teams
[**ApiV1OrganizationsOrgIdTeamsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsGet) | **Get** /api/v1/organizations/{org_id}/teams | List all teams in an organization
[**ApiV1OrganizationsOrgIdTeamsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsPost) | **Post** /api/v1/organizations/{org_id}/teams | Create a new team
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres | List all PostgresAddons for a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres/{id}/actions/backup | Trigger an immediate backup
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres/{id}/actions/fence | Fence or unfence the PostgreSQL cluster
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres/{id}/actions/hibernate | Hibernate or wake the PostgreSQL cluster
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres/{id}/backups | List backups for a PostgresAddon
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres/{id}/credentials/{database} | Get JIT credentials for a PostgresAddon database
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres/{id} | Delete a PostgresAddon
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres/{id} | Get a specific PostgresAddon
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut) | **Put** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres/{id} | Update a PostgresAddon
[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/addons/postgres | Create a new PostgresAddon
[**ApiV1OrganizationsOrgIdTeamsTeamNameDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name} | Delete a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name} | Get a specific team by name
[**ApiV1OrganizationsOrgIdTeamsTeamNameMembersGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameMembersGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/members | List members of a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/members/{id} | Remove a member from a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPut) | **Put** /api/v1/organizations/{org_id}/teams/{team_name}/members/{id} | Update a team member&#39;s role
[**ApiV1OrganizationsOrgIdTeamsTeamNameMembersPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameMembersPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/members | Add a member to a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/object-stores | List all ObjectStores for a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/object-stores/{id} | Delete an ObjectStore
[**ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/object-stores/{id} | Get a specific ObjectStore
[**ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut) | **Put** /api/v1/organizations/{org_id}/teams/{team_name}/object-stores/{id} | Update an ObjectStore
[**ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/object-stores | Add a new ObjectStore for backup storage
[**ApiV1OrganizationsOrgIdTeamsTeamNamePut**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNamePut) | **Put** /api/v1/organizations/{org_id}/teams/{team_name} | Update a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/secrets | List all secrets for a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/secrets/{id} | Delete a secret
[**ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/secrets/{id} | Get a specific secret
[**ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPut) | **Put** /api/v1/organizations/{org_id}/teams/{team_name}/secrets/{id} | Update a secret
[**ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/secrets | Create a new secret
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks | List all stacks for a team
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/builds/{build_id} | Get a specific build under a stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/builds | List all builds under a stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/connections/{connection_id} | Delete stack connection
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut) | **Put** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/connections/{connection_id} | Update stack connection
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/connections | List stack connections
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/connections | Create stack connection
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id} | Delete a stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id} | Get a specific stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/logs | Get logs for a stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/metrics | Get metrics for a stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut) | **Put** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id} | Update a stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/resources | List all stack resources under a stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/resources/{resource_name}/actions/restart | Restart a stack resource
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/resources/{resource_name}/builds | List all builds for a stack resource
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/resources/{resource_name} | Get a specific stack resource by name
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/resources/{resource_name}/logs | Get logs for a stack resource
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/resources/{resource_name}/metrics | Get metrics for a stack resource
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/topology | Get stack topology
[**ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/stacks | Create a new stack
[**ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/volumes/{id} | Delete a volume
[**ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/volumes/{id} | Get a specific volume
[**ApiV1OrganizationsOrgIdTeamsTeamNameVolumesPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameVolumesPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/volumes | Create a new volume
[**ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/workspace-users/current | Get the workspace user for the current user
[**ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdDelete) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/workspace-users/{id} | Delete a workspace user
[**ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGet) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/workspace-users/{id} | Get a workspace user by ID
[**ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPut) | **Put** /api/v1/organizations/{org_id}/teams/{team_name}/workspace-users/{id} | Update a workspace user
[**ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPost**](DefaultApi.md#ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPost) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/workspace-users | Create a new workspace user
[**ApiV1OrganizationsOrgIdUsersGet**](DefaultApi.md#ApiV1OrganizationsOrgIdUsersGet) | **Get** /api/v1/organizations/{org_id}/users | List all users in an organization
[**ApiV1TeamRolesGet**](DefaultApi.md#ApiV1TeamRolesGet) | **Get** /api/v1/team-roles | List available team membership roles
[**ApiV1UserSignupPost**](DefaultApi.md#ApiV1UserSignupPost) | **Post** /api/v1/user-signup | Create new user
[**ApiV1UsersCurrentGet**](DefaultApi.md#ApiV1UsersCurrentGet) | **Get** /api/v1/users/current | Get the current authenticated user
[**ApiV1UsersCurrentTeamsGet**](DefaultApi.md#ApiV1UsersCurrentTeamsGet) | **Get** /api/v1/users/current/teams | List teams for the current authenticated user
[**ApiV1UsersIdGet**](DefaultApi.md#ApiV1UsersIdGet) | **Get** /api/v1/users/{id} | Get a user



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
    demoteAdminRequest := *openapiclient.NewDemoteAdminRequest("TeamName_example") // DemoteAdminRequest | 

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
    orgInviteCreateRequest := *openapiclient.NewOrgInviteCreateRequest("Email_example", "TeamName_example", "Role_example", int32(123)) // OrgInviteCreateRequest | 

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

List all object stores the user has access to across all teams



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


## ApiV1OrganizationsOrgIdSecretsGet

> SecretList ApiV1OrganizationsOrgIdSecretsGet(ctx, orgId).Execute()

List all secrets the user has access to across all teams



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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdSecretsGet(context.Background(), orgId).Execute()
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

List all stacks the user has access to across all teams



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


## ApiV1OrganizationsOrgIdTeamsGet

> TeamList ApiV1OrganizationsOrgIdTeamsGet(ctx, orgId).Execute()

List all teams in an organization

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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsGet`: TeamList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**TeamList**](TeamList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsPost

> Team ApiV1OrganizationsOrgIdTeamsPost(ctx, orgId).TeamCreateRequest(teamCreateRequest).Execute()

Create a new team

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
    teamCreateRequest := *openapiclient.NewTeamCreateRequest("Name_example") // TeamCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsPost(context.Background(), orgId).TeamCreateRequest(teamCreateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsPost`: Team
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **teamCreateRequest** | [**TeamCreateRequest**](TeamCreateRequest.md) |  | 

### Return type

[**Team**](Team.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet

> PostgresAddonList ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet(ctx, orgId, teamName).Execute()

List all PostgresAddons for a team

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
    teamName := "teamName_example" // string | The name of the team

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet(context.Background(), orgId, teamName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet`: PostgresAddonList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost

> ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost202Response ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost(ctx, orgId, teamName, id).ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest(apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest := *openapiclient.NewApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest() // ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost(context.Background(), orgId, teamName, id).ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest(apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost`: ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost202Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest** | [**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest**](ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest.md) |  | 

### Return type

[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost202Response**](ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost202Response.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost

> ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost200Response ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost(ctx, orgId, teamName, id).ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest(apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest := *openapiclient.NewApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest(false) // ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost(context.Background(), orgId, teamName, id).ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest(apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost`: ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost200Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest** | [**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest**](ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest.md) |  | 

### Return type

[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost200Response**](ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost200Response.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost

> ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost200Response ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost(ctx, orgId, teamName, id).ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest(apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest := *openapiclient.NewApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest(false) // ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost(context.Background(), orgId, teamName, id).ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest(apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost`: ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost200Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **apiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest** | [**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest**](ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest.md) |  | 

### Return type

[**ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost200Response**](ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost200Response.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGet

> PostgresBackupList ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGet(ctx, orgId, teamName, id).Limit(limit).Offset(offset).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    limit := int32(56) // int32 |  (optional) (default to 20)
    offset := int32(56) // int32 |  (optional) (default to 0)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGet(context.Background(), orgId, teamName, id).Limit(limit).Offset(offset).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGet`: PostgresBackupList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdBackupsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGet

> PostgresCredentials ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGet(ctx, orgId, teamName, id, database).Superuser(superuser).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    database := "database_example" // string | The name of the database
    superuser := true // bool | Whether to return superuser credentials (optional) (default to false)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGet(context.Background(), orgId, teamName, id, database).Superuser(superuser).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGet`: PostgresCredentials
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**database** | **string** | The name of the database | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdCredentialsDatabaseGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete

> PostgresAddon ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete`: PostgresAddon
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet

> PostgresAddon ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet`: PostgresAddon
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut

> PostgresAddon ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut(ctx, orgId, teamName, id).PostgresAddon(postgresAddon).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    postgresAddon := *openapiclient.NewPostgresAddon("Name_example", *openapiclient.NewPostgresAddonSpec(*openapiclient.NewPostgresVersion(int32(123)), *openapiclient.NewPostgresInstances(int32(123)), *openapiclient.NewPostgresStorage("Size_example", "StorageClass_example"))) // PostgresAddon | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut(context.Background(), orgId, teamName, id).PostgresAddon(postgresAddon).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut`: PostgresAddon
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost

> PostgresAddon ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost(ctx, orgId, teamName).PostgresAddon(postgresAddon).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    postgresAddon := *openapiclient.NewPostgresAddon("Name_example", *openapiclient.NewPostgresAddonSpec(*openapiclient.NewPostgresVersion(int32(123)), *openapiclient.NewPostgresInstances(int32(123)), *openapiclient.NewPostgresStorage("Size_example", "StorageClass_example"))) // PostgresAddon | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost(context.Background(), orgId, teamName).PostgresAddon(postgresAddon).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost`: PostgresAddon
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPostRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameDelete

> ApiV1OrganizationsOrgIdTeamsTeamNameDelete(ctx, orgId, teamName).Execute()

Delete a team

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
    teamName := "teamName_example" // string | The name of the team

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameDelete(context.Background(), orgId, teamName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameGet

> Team ApiV1OrganizationsOrgIdTeamsTeamNameGet(ctx, orgId, teamName).Execute()

Get a specific team by name

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
    teamName := "teamName_example" // string | The name of the team

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameGet(context.Background(), orgId, teamName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameGet`: Team
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**Team**](Team.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameMembersGet

> TeamMembershipList ApiV1OrganizationsOrgIdTeamsTeamNameMembersGet(ctx, orgId, teamName).Execute()

List members of a team

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
    teamName := "teamName_example" // string | The name of the team

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersGet(context.Background(), orgId, teamName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameMembersGet`: TeamMembershipList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameMembersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**TeamMembershipList**](TeamMembershipList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdDelete

> ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdDelete(ctx, orgId, teamName, id).Execute()

Remove a member from a team

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdDelete(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameMembersIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPut

> TeamMembership ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPut(ctx, orgId, teamName, id).UpdateTeamMemberRoleRequest(updateTeamMemberRoleRequest).Execute()

Update a team member's role

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    updateTeamMemberRoleRequest := *openapiclient.NewUpdateTeamMemberRoleRequest(openapiclient.TeamRole("Developer")) // UpdateTeamMemberRoleRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPut(context.Background(), orgId, teamName, id).UpdateTeamMemberRoleRequest(updateTeamMemberRoleRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPut`: TeamMembership
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameMembersIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **updateTeamMemberRoleRequest** | [**UpdateTeamMemberRoleRequest**](UpdateTeamMemberRoleRequest.md) |  | 

### Return type

[**TeamMembership**](TeamMembership.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameMembersPost

> TeamMembership ApiV1OrganizationsOrgIdTeamsTeamNameMembersPost(ctx, orgId, teamName).AddTeamMemberRequest(addTeamMemberRequest).Execute()

Add a member to a team

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
    teamName := "teamName_example" // string | The name of the team
    addTeamMemberRequest := *openapiclient.NewAddTeamMemberRequest("UserId_example", openapiclient.TeamRole("Developer")) // AddTeamMemberRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersPost(context.Background(), orgId, teamName).AddTeamMemberRequest(addTeamMemberRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameMembersPost`: TeamMembership
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameMembersPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameMembersPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **addTeamMemberRequest** | [**AddTeamMemberRequest**](AddTeamMemberRequest.md) |  | 

### Return type

[**TeamMembership**](TeamMembership.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet

> ObjectStoreList ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet(ctx, orgId, teamName).Execute()

List all ObjectStores for a team

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
    teamName := "teamName_example" // string | The name of the team

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet(context.Background(), orgId, teamName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet`: ObjectStoreList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdDelete

> ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdDelete(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdDelete(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet

> ObjectStore ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet`: ObjectStore
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut

> ObjectStore ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut(ctx, orgId, teamName, id).ObjectStore(objectStore).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    objectStore := *openapiclient.NewObjectStore("Name_example", *openapiclient.NewObjectStoreSpec(*openapiclient.NewObjectStoreConfiguration(), "DestinationPath_example")) // ObjectStore | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut(context.Background(), orgId, teamName, id).ObjectStore(objectStore).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut`: ObjectStore
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost

> ObjectStore ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost(ctx, orgId, teamName).ObjectStore(objectStore).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    objectStore := *openapiclient.NewObjectStore("Name_example", *openapiclient.NewObjectStoreSpec(*openapiclient.NewObjectStoreConfiguration(), "DestinationPath_example")) // ObjectStore | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost(context.Background(), orgId, teamName).ObjectStore(objectStore).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost`: ObjectStore
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPostRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNamePut

> Team ApiV1OrganizationsOrgIdTeamsTeamNamePut(ctx, orgId, teamName).TeamUpdateRequest(teamUpdateRequest).Execute()

Update a team

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
    teamName := "teamName_example" // string | The name of the team
    teamUpdateRequest := *openapiclient.NewTeamUpdateRequest("Name_example") // TeamUpdateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNamePut(context.Background(), orgId, teamName).TeamUpdateRequest(teamUpdateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNamePut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNamePut`: Team
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNamePut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNamePutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **teamUpdateRequest** | [**TeamUpdateRequest**](TeamUpdateRequest.md) |  | 

### Return type

[**Team**](Team.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet

> SecretList ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet(ctx, orgId, teamName).Execute()

List all secrets for a team

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
    teamName := "teamName_example" // string | The name of the team

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet(context.Background(), orgId, teamName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet`: SecretList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameSecretsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



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


## ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDelete

> ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDelete(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDelete(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet

> Secret ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet`: Secret
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPut

> Secret ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPut(ctx, orgId, teamName, id).Secret(secret).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    secret := *openapiclient.NewSecret("Name_example", openapiclient.SecretType("Generic"), []openapiclient.SecretData{*openapiclient.NewSecretData("Key_example", "Value_example")}) // Secret | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPut(context.Background(), orgId, teamName, id).Secret(secret).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPut`: Secret
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost

> Secret ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost(ctx, orgId, teamName).Secret(secret).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    secret := *openapiclient.NewSecret("Name_example", openapiclient.SecretType("Generic"), []openapiclient.SecretData{*openapiclient.NewSecretData("Key_example", "Value_example")}) // Secret | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost(context.Background(), orgId, teamName).Secret(secret).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost`: Secret
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameSecretsPostRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet

> StackList ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet(ctx, orgId, teamName).Limit(limit).Offset(offset).Execute()

List all stacks for a team

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
    teamName := "teamName_example" // string | The name of the team
    limit := int32(56) // int32 |  (optional) (default to 20)
    offset := int32(56) // int32 |  (optional) (default to 0)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet(context.Background(), orgId, teamName).Limit(limit).Offset(offset).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet`: StackList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGet

> ImageBuild ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGet(ctx, orgId, teamName, id, buildId).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    buildId := "buildId_example" // string | The ID of the build

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGet(context.Background(), orgId, teamName, id, buildId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGet`: ImageBuild
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**buildId** | **string** | The ID of the build | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsBuildIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGet

> ImageBuildList ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGet`: ImageBuildList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdBuildsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete

> ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete(ctx, orgId, teamName, id, connectionId).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    connectionId := "connectionId_example" // string | The ID of the stack connection

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete(context.Background(), orgId, teamName, id, connectionId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**connectionId** | **string** | The ID of the stack connection | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut

> StackConnection ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut(ctx, orgId, teamName, id, connectionId).StackConnection(stackConnection).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    connectionId := "connectionId_example" // string | The ID of the stack connection
    stackConnection := *openapiclient.NewStackConnection("Kind_example", *openapiclient.NewTopologyNodeRef("Type_example"), *openapiclient.NewTopologyNodeRef("Type_example")) // StackConnection | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut(context.Background(), orgId, teamName, id, connectionId).StackConnection(stackConnection).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut`: StackConnection
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**connectionId** | **string** | The ID of the stack connection | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet

> StackConnectionList ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet`: StackConnectionList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost

> StackConnection ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost(ctx, orgId, teamName, id).StackConnection(stackConnection).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    stackConnection := *openapiclient.NewStackConnection("Kind_example", *openapiclient.NewTopologyNodeRef("Type_example"), *openapiclient.NewTopologyNodeRef("Type_example")) // StackConnection | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost(context.Background(), orgId, teamName, id).StackConnection(stackConnection).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost`: StackConnection
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPostRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete

> Stack ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet

> Stack ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGet

> *os.File ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGet(ctx, orgId, teamName, id).Follow(follow).Tail(tail).Since(since).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    follow := true // bool |  (optional) (default to false)
    tail := int32(56) // int32 |  (optional) (default to 100)
    since := "since_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGet(context.Background(), orgId, teamName, id).Follow(follow).Tail(tail).Since(since).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGet`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdLogsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGet

> *os.File ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGet(ctx, orgId, teamName, id).Stream(stream).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    stream := true // bool |  (optional) (default to false)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGet(context.Background(), orgId, teamName, id).Stream(stream).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGet`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdMetricsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut

> Stack ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut(ctx, orgId, teamName, id).Stack(stack).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    stack := *openapiclient.NewStack("Name_example", *openapiclient.NewStackSpec([]openapiclient.StackResource{*openapiclient.NewStackResource("Name_example")})) // Stack | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut(context.Background(), orgId, teamName, id).Stack(stack).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGet

> StackResourceList ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGet`: StackResourceList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPost

> StackResource ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPost(ctx, orgId, teamName, id, resourceName).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPost(context.Background(), orgId, teamName, id, resourceName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPost`: StackResource
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameActionsRestartPostRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGet

> ImageBuildList ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGet(ctx, orgId, teamName, id, resourceName).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGet(context.Background(), orgId, teamName, id, resourceName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGet`: ImageBuildList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameBuildsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGet

> StackResource ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGet(ctx, orgId, teamName, id, resourceName).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGet(context.Background(), orgId, teamName, id, resourceName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGet`: StackResource
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGet

> *os.File ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGet(ctx, orgId, teamName, id, resourceName).Follow(follow).Tail(tail).Since(since).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource
    follow := true // bool |  (optional) (default to false)
    tail := int32(56) // int32 |  (optional) (default to 100)
    since := "since_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGet(context.Background(), orgId, teamName, id, resourceName).Follow(follow).Tail(tail).Since(since).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGet`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameLogsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGet

> ResourceMetrics ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGet(ctx, orgId, teamName, id, resourceName).Stream(stream).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    resourceName := "resourceName_example" // string | The name of the stack resource
    stream := true // bool | Whether to stream metrics via Server-Sent Events (SSE) (optional) (default to false)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGet(context.Background(), orgId, teamName, id, resourceName).Stream(stream).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGet`: ResourceMetrics
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**resourceName** | **string** | The name of the stack resource | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdResourcesResourceNameMetricsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet

> StackTopology ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet`: StackTopology
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost

> Stack ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost(ctx, orgId, teamName).Stack(stack).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    stack := *openapiclient.NewStack("Name_example", *openapiclient.NewStackSpec([]openapiclient.StackResource{*openapiclient.NewStackResource("Name_example")})) // Stack | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost(context.Background(), orgId, teamName).Stack(stack).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameStacksPostRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdDelete

> ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdDelete(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdDelete(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGet

> Volume ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGet`: Volume
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameVolumesIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameVolumesPost

> Volume ApiV1OrganizationsOrgIdTeamsTeamNameVolumesPost(ctx, orgId, teamName).Volume(volume).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    volume := *openapiclient.NewVolume("Name_example", *openapiclient.NewVolumeSpec("Size_example", false, openapiclient.VolumeAccessMode("ReadWriteOnce"))) // Volume | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameVolumesPost(context.Background(), orgId, teamName).Volume(volume).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameVolumesPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameVolumesPost`: Volume
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameVolumesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameVolumesPostRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGet

> WorkspaceUser ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGet(ctx, orgId, teamName).Execute()

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
    teamName := "teamName_example" // string | The name of the team

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGet(context.Background(), orgId, teamName).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGet`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersCurrentGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdDelete

> ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdDelete(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdDelete(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGet

> WorkspaceUser ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGet(ctx, orgId, teamName, id).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGet(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGet`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPut

> WorkspaceUser ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPut(ctx, orgId, teamName, id).WorkspaceUser(workspaceUser).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    id := "id_example" // string | The id of record
    workspaceUser := *openapiclient.NewWorkspaceUser([]string{"Workspaces_example"}) // WorkspaceUser | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPut(context.Background(), orgId, teamName, id).WorkspaceUser(workspaceUser).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPut`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPost

> WorkspaceUser ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPost(ctx, orgId, teamName).WorkspaceUser(workspaceUser).Execute()

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
    teamName := "teamName_example" // string | The name of the team
    workspaceUser := *openapiclient.NewWorkspaceUser([]string{"Workspaces_example"}) // WorkspaceUser | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPost(context.Background(), orgId, teamName).WorkspaceUser(workspaceUser).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPost`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdTeamsTeamNameWorkspaceUsersPostRequest struct via the builder pattern


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


## ApiV1TeamRolesGet

> TeamRoleList ApiV1TeamRolesGet(ctx).Execute()

List available team membership roles



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
    resp, r, err := apiClient.DefaultApi.ApiV1TeamRolesGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1TeamRolesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1TeamRolesGet`: TeamRoleList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1TeamRolesGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1TeamRolesGetRequest struct via the builder pattern


### Return type

[**TeamRoleList**](TeamRoleList.md)

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


## ApiV1UsersCurrentTeamsGet

> TeamList ApiV1UsersCurrentTeamsGet(ctx).Execute()

List teams for the current authenticated user

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
    resp, r, err := apiClient.DefaultApi.ApiV1UsersCurrentTeamsGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1UsersCurrentTeamsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1UsersCurrentTeamsGet`: TeamList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1UsersCurrentTeamsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1UsersCurrentTeamsGetRequest struct via the builder pattern


### Return type

[**TeamList**](TeamList.md)

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

