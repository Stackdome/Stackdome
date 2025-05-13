# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ApiV1AuthLoginPost**](DefaultApi.md#ApiV1AuthLoginPost) | **Post** /api/v1/auth/login | User login
[**ApiV1OrganizationsDefaultGet**](DefaultApi.md#ApiV1OrganizationsDefaultGet) | **Get** /api/v1/organizations/default | Get the default organization
[**ApiV1OrganizationsIdClustersGet**](DefaultApi.md#ApiV1OrganizationsIdClustersGet) | **Get** /api/v1/organizations/{id}/clusters | List all clusters for an organization
[**ApiV1OrganizationsIdClustersPost**](DefaultApi.md#ApiV1OrganizationsIdClustersPost) | **Post** /api/v1/organizations/{id}/clusters | Add a new cluster
[**ApiV1OrganizationsIdGet**](DefaultApi.md#ApiV1OrganizationsIdGet) | **Get** /api/v1/organizations/{id} | Get an organization
[**ApiV1OrganizationsIdPut**](DefaultApi.md#ApiV1OrganizationsIdPut) | **Put** /api/v1/organizations/{id} | Update an organization
[**ApiV1OrganizationsIdRemoteSyncServersGet**](DefaultApi.md#ApiV1OrganizationsIdRemoteSyncServersGet) | **Get** /api/v1/organizations/{id}/remote-sync-servers | List all RemoteSyncServer objects for an organization
[**ApiV1OrganizationsIdRemoteSyncServersPost**](DefaultApi.md#ApiV1OrganizationsIdRemoteSyncServersPost) | **Post** /api/v1/organizations/{id}/remote-sync-servers | Create a new remote sync server
[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet) | **Get** /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries | List all image registries for a cluster
[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete) | **Delete** /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries/{id} | Delete an image registry object
[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdGet) | **Get** /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries/{id} | Get a specific image registry object
[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesPost) | **Post** /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries | Create a new image registry
[**ApiV1OrganizationsOrgIdClustersIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersIdDelete) | **Delete** /api/v1/organizations/{org_id}/clusters/{id} | Delete a cluster object
[**ApiV1OrganizationsOrgIdClustersIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersIdGet) | **Get** /api/v1/organizations/{org_id}/clusters/{id} | Get a specific cluster object
[**ApiV1OrganizationsOrgIdClustersIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdClustersIdPut) | **Put** /api/v1/organizations/{org_id}/clusters/{id} | Update a cluster object
[**ApiV1OrganizationsOrgIdRemoteSyncServersCurrentGet**](DefaultApi.md#ApiV1OrganizationsOrgIdRemoteSyncServersCurrentGet) | **Get** /api/v1/organizations/{org_id}/remote-sync-servers/current | Get RemoteSyncServer for the current user
[**ApiV1OrganizationsOrgIdRemoteSyncServersIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdRemoteSyncServersIdDelete) | **Delete** /api/v1/organizations/{org_id}/remote-sync-servers/{id} | Delete a RemoteSyncServer object
[**ApiV1OrganizationsOrgIdRemoteSyncServersIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdRemoteSyncServersIdGet) | **Get** /api/v1/organizations/{org_id}/remote-sync-servers/{id} | Get a specific RemoteSyncServer object
[**ApiV1OrganizationsOrgIdRemoteSyncServersIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdRemoteSyncServersIdPut) | **Put** /api/v1/organizations/{org_id}/remote-sync-servers/{id} | Update a RemoteSyncServer object
[**ApiV1OrganizationsOrgIdStacksCurrentGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksCurrentGet) | **Get** /api/v1/organizations/{org_id}/stacks/current | List all stacks of the current user
[**ApiV1OrganizationsOrgIdStacksGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksGet) | **Get** /api/v1/organizations/{org_id}/stacks | List all Stacks for an organization
[**ApiV1OrganizationsOrgIdStacksIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksIdDelete) | **Delete** /api/v1/organizations/{org_id}/stacks/{id} | Delete a Stack
[**ApiV1OrganizationsOrgIdStacksIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksIdGet) | **Get** /api/v1/organizations/{org_id}/stacks/{id} | Get a specific Stack
[**ApiV1OrganizationsOrgIdStacksIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksIdPut) | **Put** /api/v1/organizations/{org_id}/stacks/{id} | Update a Stack
[**ApiV1OrganizationsOrgIdStacksPost**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksPost) | **Post** /api/v1/organizations/{org_id}/stacks | Create a new stack
[**ApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGet) | **Get** /api/v1/organizations/{org_id}/stacks/{stack_id}/builds/{build_id} | Get a specific build under a stack
[**ApiV1OrganizationsOrgIdStacksStackIdBuildsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksStackIdBuildsGet) | **Get** /api/v1/organizations/{org_id}/stacks/{stack_id}/builds | List all builds under a stack
[**ApiV1OrganizationsOrgIdStacksStackIdResourcesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksStackIdResourcesGet) | **Get** /api/v1/organizations/{org_id}/stacks/{stack_id}/resources | List all StackResources under a Stack
[**ApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGet**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGet) | **Get** /api/v1/organizations/{org_id}/stacks/{stack_id}/resources/{id}/builds | List all builds for a StackResource
[**ApiV1OrganizationsOrgIdStacksStackIdResourcesIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdStacksStackIdResourcesIdPut) | **Put** /api/v1/organizations/{org_id}/stacks/{stack_id}/resources/{id} | Update a StackResource
[**ApiV1OrganizationsOrgIdVolumesCurrentGet**](DefaultApi.md#ApiV1OrganizationsOrgIdVolumesCurrentGet) | **Get** /api/v1/organizations/{org_id}/volumes/current | List of volumes for the current user
[**ApiV1OrganizationsOrgIdVolumesGet**](DefaultApi.md#ApiV1OrganizationsOrgIdVolumesGet) | **Get** /api/v1/organizations/{org_id}/volumes | List all volumes in an organization.
[**ApiV1OrganizationsOrgIdVolumesIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdVolumesIdDelete) | **Delete** /api/v1/organizations/{org_id}/volumes/{id} | Delete a volume
[**ApiV1OrganizationsOrgIdVolumesIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdVolumesIdGet) | **Get** /api/v1/organizations/{org_id}/volumes/{id} | Get a specific volume
[**ApiV1OrganizationsOrgIdVolumesIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdVolumesIdPut) | **Put** /api/v1/organizations/{org_id}/volumes/{id} | Update a volume
[**ApiV1OrganizationsOrgIdVolumesPost**](DefaultApi.md#ApiV1OrganizationsOrgIdVolumesPost) | **Post** /api/v1/organizations/{org_id}/volumes | Create a new volume
[**ApiV1OrganizationsPost**](DefaultApi.md#ApiV1OrganizationsPost) | **Post** /api/v1/organizations | Create a new organization
[**ApiV1UserSignupPost**](DefaultApi.md#ApiV1UserSignupPost) | **Post** /api/v1/user-signup | Create new user
[**ApiV1UsersCurrentGet**](DefaultApi.md#ApiV1UsersCurrentGet) | **Get** /api/v1/users/current | Get a the current authenticated user
[**ApiV1UsersIdGet**](DefaultApi.md#ApiV1UsersIdGet) | **Get** /api/v1/users/{id} | Get a user
[**ApiV1WorkspaceUsersCurrentGet**](DefaultApi.md#ApiV1WorkspaceUsersCurrentGet) | **Get** /api/v1/workspace-users/current | Get the workspace user object for the current user
[**ApiV1WorkspaceUsersIdDelete**](DefaultApi.md#ApiV1WorkspaceUsersIdDelete) | **Delete** /api/v1/workspace-users/{id} | Delete a WorkspaceUser
[**ApiV1WorkspaceUsersIdGet**](DefaultApi.md#ApiV1WorkspaceUsersIdGet) | **Get** /api/v1/workspace-users/{id} | Get a workspace user object by ID
[**ApiV1WorkspaceUsersIdPut**](DefaultApi.md#ApiV1WorkspaceUsersIdPut) | **Put** /api/v1/workspace-users/{id} | Update a WorkspaceUser
[**ApiV1WorkspaceUsersPost**](DefaultApi.md#ApiV1WorkspaceUsersPost) | **Post** /api/v1/workspace-users | Create a new workspace user object.



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


## ApiV1OrganizationsDefaultGet

> Organisation ApiV1OrganizationsDefaultGet(ctx).Execute()

Get the default organization



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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsDefaultGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsDefaultGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsDefaultGet`: Organisation
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsDefaultGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsDefaultGetRequest struct via the builder pattern


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


## ApiV1OrganizationsIdClustersGet

> ClusterList ApiV1OrganizationsIdClustersGet(ctx, id).Execute()

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
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsIdClustersGet(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsIdClustersGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsIdClustersGet`: ClusterList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsIdClustersGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsIdClustersGetRequest struct via the builder pattern


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


## ApiV1OrganizationsIdClustersPost

> Cluster ApiV1OrganizationsIdClustersPost(ctx, id).Cluster(cluster).Execute()

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
    id := "id_example" // string | The id of record
    cluster := *openapiclient.NewCluster("Name_example", "ClusterUrl_example", "ClusterCaData_example", "ClusterSaToken_example") // Cluster | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsIdClustersPost(context.Background(), id).Cluster(cluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsIdClustersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsIdClustersPost`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsIdClustersPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsIdClustersPostRequest struct via the builder pattern


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


## ApiV1OrganizationsIdRemoteSyncServersGet

> ApiV1OrganizationsIdRemoteSyncServersGet200Response ApiV1OrganizationsIdRemoteSyncServersGet(ctx, id).Execute()

List all RemoteSyncServer objects for an organization

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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsIdRemoteSyncServersGet(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsIdRemoteSyncServersGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsIdRemoteSyncServersGet`: ApiV1OrganizationsIdRemoteSyncServersGet200Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsIdRemoteSyncServersGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsIdRemoteSyncServersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**ApiV1OrganizationsIdRemoteSyncServersGet200Response**](ApiV1OrganizationsIdRemoteSyncServersGet200Response.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsIdRemoteSyncServersPost

> RemoteSyncServer ApiV1OrganizationsIdRemoteSyncServersPost(ctx, id).RemoteSyncServer(remoteSyncServer).Execute()

Create a new remote sync server

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
    remoteSyncServer := *openapiclient.NewRemoteSyncServer("Name_example", *openapiclient.NewRemoteSyncServerSpec()) // RemoteSyncServer | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsIdRemoteSyncServersPost(context.Background(), id).RemoteSyncServer(remoteSyncServer).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsIdRemoteSyncServersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsIdRemoteSyncServersPost`: RemoteSyncServer
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsIdRemoteSyncServersPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsIdRemoteSyncServersPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **remoteSyncServer** | [**RemoteSyncServer**](RemoteSyncServer.md) |  | 

### Return type

[**RemoteSyncServer**](RemoteSyncServer.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet

> ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet200Response ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet(ctx, orgId, clusterId).Execute()

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
    orgId := "orgId_example" // string | 
    clusterId := "clusterId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet(context.Background(), orgId, clusterId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet`: ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet200Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**clusterId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet200Response**](ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesGet200Response.md)

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

Delete an image registry object

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
    orgId := "orgId_example" // string | 
    clusterId := "clusterId_example" // string | 
    id := "id_example" // string | 

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
**orgId** | **string** |  | 
**clusterId** | **string** |  | 
**id** | **string** |  | 

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

Get a specific image registry object

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
    orgId := "orgId_example" // string | 
    clusterId := "clusterId_example" // string | 
    id := "id_example" // string | 

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
**orgId** | **string** |  | 
**clusterId** | **string** |  | 
**id** | **string** |  | 

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
    orgId := "orgId_example" // string | 
    clusterId := "clusterId_example" // string | 
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
**orgId** | **string** |  | 
**clusterId** | **string** |  | 

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


## ApiV1OrganizationsOrgIdClustersIdDelete

> ApiV1OrganizationsOrgIdClustersIdDelete(ctx, orgId, id).Execute()

Delete a cluster object

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 

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
**orgId** | **string** |  | 
**id** | **string** |  | 

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

Get a specific cluster object

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 

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
**orgId** | **string** |  | 
**id** | **string** |  | 

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


## ApiV1OrganizationsOrgIdClustersIdPut

> Cluster ApiV1OrganizationsOrgIdClustersIdPut(ctx, orgId, id).Cluster(cluster).Execute()

Update a cluster object

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 
    cluster := *openapiclient.NewCluster("Name_example", "ClusterUrl_example", "ClusterCaData_example", "ClusterSaToken_example") // Cluster | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdClustersIdPut(context.Background(), orgId, id).Cluster(cluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdClustersIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdClustersIdPut`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdClustersIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdClustersIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdRemoteSyncServersCurrentGet

> RemoteSyncServer ApiV1OrganizationsOrgIdRemoteSyncServersCurrentGet(ctx, orgId).Execute()

Get RemoteSyncServer for the current user

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
    orgId := "orgId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersCurrentGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersCurrentGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdRemoteSyncServersCurrentGet`: RemoteSyncServer
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersCurrentGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRemoteSyncServersCurrentGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**RemoteSyncServer**](RemoteSyncServer.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdRemoteSyncServersIdDelete

> ApiV1OrganizationsOrgIdRemoteSyncServersIdDelete(ctx, orgId, id).Execute()

Delete a RemoteSyncServer object

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersIdDelete(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRemoteSyncServersIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdRemoteSyncServersIdGet

> RemoteSyncServer ApiV1OrganizationsOrgIdRemoteSyncServersIdGet(ctx, orgId, id).Execute()

Get a specific RemoteSyncServer object

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersIdGet(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdRemoteSyncServersIdGet`: RemoteSyncServer
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRemoteSyncServersIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**RemoteSyncServer**](RemoteSyncServer.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdRemoteSyncServersIdPut

> RemoteSyncServer ApiV1OrganizationsOrgIdRemoteSyncServersIdPut(ctx, orgId, id).RemoteSyncServer(remoteSyncServer).Execute()

Update a RemoteSyncServer object

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 
    remoteSyncServer := *openapiclient.NewRemoteSyncServer("Name_example", *openapiclient.NewRemoteSyncServerSpec()) // RemoteSyncServer | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersIdPut(context.Background(), orgId, id).RemoteSyncServer(remoteSyncServer).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdRemoteSyncServersIdPut`: RemoteSyncServer
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdRemoteSyncServersIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdRemoteSyncServersIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **remoteSyncServer** | [**RemoteSyncServer**](RemoteSyncServer.md) |  | 

### Return type

[**RemoteSyncServer**](RemoteSyncServer.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdStacksCurrentGet

> StackList ApiV1OrganizationsOrgIdStacksCurrentGet(ctx, orgId).Execute()

List all stacks of the current user

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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksCurrentGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksCurrentGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksCurrentGet`: StackList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksCurrentGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksCurrentGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksGet

> StackList ApiV1OrganizationsOrgIdStacksGet(ctx, orgId).Limit(limit).Offset(offset).Execute()

List all Stacks for an organization

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
    limit := int32(56) // int32 |  (optional) (default to 20)
    offset := int32(56) // int32 |  (optional) (default to 0)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksGet(context.Background(), orgId).Limit(limit).Offset(offset).Execute()
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


## ApiV1OrganizationsOrgIdStacksIdDelete

> ApiV1OrganizationsOrgIdStacksIdDelete(ctx, orgId, id).Execute()

Delete a Stack

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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksIdDelete(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksIdDelete``: %v\n", err)
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

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksIdGet

> Stack ApiV1OrganizationsOrgIdStacksIdGet(ctx, orgId, id).Execute()

Get a specific Stack

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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksIdGet(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksIdGet`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksIdPut

> Stack ApiV1OrganizationsOrgIdStacksIdPut(ctx, orgId, id).Stack(stack).Execute()

Update a Stack

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
    stack := *openapiclient.NewStack("Name_example", "WorkspaceName_example", *openapiclient.NewStackSpec([]openapiclient.StackResource{*openapiclient.NewStackResource("Name_example")})) // Stack | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksIdPut(context.Background(), orgId, id).Stack(stack).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksIdPut`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksPost

> Stack ApiV1OrganizationsOrgIdStacksPost(ctx, orgId).Stack(stack).Execute()

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
    stack := *openapiclient.NewStack("Name_example", "WorkspaceName_example", *openapiclient.NewStackSpec([]openapiclient.StackResource{*openapiclient.NewStackResource("Name_example")})) // Stack | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksPost(context.Background(), orgId).Stack(stack).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksPost`: Stack
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksPostRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGet

> ImageBuild ApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGet(ctx, orgId, stackId, buildId).Execute()

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
    stackId := "stackId_example" // string | The ID of the stack
    buildId := "buildId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGet(context.Background(), orgId, stackId, buildId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGet`: ImageBuild
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**stackId** | **string** | The ID of the stack | 
**buildId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksStackIdBuildsBuildIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksStackIdBuildsGet

> ImageBuildList ApiV1OrganizationsOrgIdStacksStackIdBuildsGet(ctx, orgId, stackId).Execute()

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
    stackId := "stackId_example" // string | The ID of the stack

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdBuildsGet(context.Background(), orgId, stackId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdBuildsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksStackIdBuildsGet`: ImageBuildList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdBuildsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**stackId** | **string** | The ID of the stack | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksStackIdBuildsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksStackIdResourcesGet

> StackResourceList ApiV1OrganizationsOrgIdStacksStackIdResourcesGet(ctx, orgId, stackId).Execute()

List all StackResources under a Stack

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
    stackId := "stackId_example" // string | The ID of the stack

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesGet(context.Background(), orgId, stackId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksStackIdResourcesGet`: StackResourceList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**stackId** | **string** | The ID of the stack | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksStackIdResourcesGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGet

> ImageBuildList ApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGet(ctx, orgId, stackId, id).Execute()

List all builds for a StackResource

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
    stackId := "stackId_example" // string | The ID of the stack
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGet(context.Background(), orgId, stackId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGet`: ImageBuildList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**stackId** | **string** | The ID of the stack | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksStackIdResourcesIdBuildsGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdStacksStackIdResourcesIdPut

> StackResource ApiV1OrganizationsOrgIdStacksStackIdResourcesIdPut(ctx, orgId, stackId, id).StackResource(stackResource).Execute()

Update a StackResource

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
    stackId := "stackId_example" // string | The ID of the stack
    id := "id_example" // string | The id of record
    stackResource := *openapiclient.NewStackResource("Name_example") // StackResource | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesIdPut(context.Background(), orgId, stackId, id).StackResource(stackResource).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdStacksStackIdResourcesIdPut`: StackResource
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdStacksStackIdResourcesIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**stackId** | **string** | The ID of the stack | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdStacksStackIdResourcesIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdVolumesCurrentGet

> VolumeList ApiV1OrganizationsOrgIdVolumesCurrentGet(ctx, orgId).Execute()

List of volumes for the current user

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
    orgId := "orgId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdVolumesCurrentGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdVolumesCurrentGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdVolumesCurrentGet`: VolumeList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdVolumesCurrentGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdVolumesCurrentGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**VolumeList**](VolumeList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdVolumesGet

> VolumeList ApiV1OrganizationsOrgIdVolumesGet(ctx, orgId).Execute()

List all volumes in an organization.

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
    orgId := "orgId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdVolumesGet(context.Background(), orgId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdVolumesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdVolumesGet`: VolumeList
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdVolumesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdVolumesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**VolumeList**](VolumeList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdVolumesIdDelete

> ApiV1OrganizationsOrgIdVolumesIdDelete(ctx, orgId, id).Execute()

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdVolumesIdDelete(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdVolumesIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdVolumesIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdVolumesIdGet

> Volume ApiV1OrganizationsOrgIdVolumesIdGet(ctx, orgId, id).Execute()

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdVolumesIdGet(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdVolumesIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdVolumesIdGet`: Volume
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdVolumesIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdVolumesIdGetRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdVolumesIdPut

> Volume ApiV1OrganizationsOrgIdVolumesIdPut(ctx, orgId, id).Volume(volume).Execute()

Update a volume

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
    orgId := "orgId_example" // string | 
    id := "id_example" // string | 
    volume := *openapiclient.NewVolume("Name_example", "WorkspaceName_example", *openapiclient.NewVolumeSpec("Size_example", false, openapiclient.VolumeAccessMode("ReadWriteOnce"))) // Volume | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdVolumesIdPut(context.Background(), orgId, id).Volume(volume).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdVolumesIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdVolumesIdPut`: Volume
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdVolumesIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdVolumesIdPutRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdVolumesPost

> Volume ApiV1OrganizationsOrgIdVolumesPost(ctx, orgId).Volume(volume).Execute()

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
    volume := *openapiclient.NewVolume("Name_example", "WorkspaceName_example", *openapiclient.NewVolumeSpec("Size_example", false, openapiclient.VolumeAccessMode("ReadWriteOnce"))) // Volume | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdVolumesPost(context.Background(), orgId).Volume(volume).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdVolumesPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdVolumesPost`: Volume
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdVolumesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdVolumesPostRequest struct via the builder pattern


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


## ApiV1OrganizationsPost

> Organisation ApiV1OrganizationsPost(ctx).Organisation(organisation).Execute()

Create a new organization



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
    organisation := *openapiclient.NewOrganisation() // Organisation | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsPost(context.Background()).Organisation(organisation).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsPost`: Organisation
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsPostRequest struct via the builder pattern


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

Get a the current authenticated user



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


## ApiV1WorkspaceUsersCurrentGet

> WorkspaceUser ApiV1WorkspaceUsersCurrentGet(ctx).Execute()

Get the workspace user object for the current user

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
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceUsersCurrentGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceUsersCurrentGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1WorkspaceUsersCurrentGet`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1WorkspaceUsersCurrentGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WorkspaceUsersCurrentGetRequest struct via the builder pattern


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


## ApiV1WorkspaceUsersIdDelete

> ApiV1WorkspaceUsersIdDelete(ctx, id).Execute()

Delete a WorkspaceUser

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
    id := "id_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceUsersIdDelete(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceUsersIdDelete``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WorkspaceUsersIdDeleteRequest struct via the builder pattern


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


## ApiV1WorkspaceUsersIdGet

> WorkspaceUser ApiV1WorkspaceUsersIdGet(ctx, id).Execute()

Get a workspace user object by ID

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
    id := "id_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceUsersIdGet(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceUsersIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1WorkspaceUsersIdGet`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1WorkspaceUsersIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WorkspaceUsersIdGetRequest struct via the builder pattern


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


## ApiV1WorkspaceUsersIdPut

> WorkspaceUser ApiV1WorkspaceUsersIdPut(ctx, id).WorkspaceUser(workspaceUser).Execute()

Update a WorkspaceUser

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
    id := "id_example" // string | 
    workspaceUser := *openapiclient.NewWorkspaceUser([]string{"Workspaces_example"}) // WorkspaceUser | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceUsersIdPut(context.Background(), id).WorkspaceUser(workspaceUser).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceUsersIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1WorkspaceUsersIdPut`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1WorkspaceUsersIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WorkspaceUsersIdPutRequest struct via the builder pattern


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


## ApiV1WorkspaceUsersPost

> WorkspaceUser ApiV1WorkspaceUsersPost(ctx).WorkspaceUser(workspaceUser).Execute()

Create a new workspace user object.

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
    workspaceUser := *openapiclient.NewWorkspaceUser([]string{"Workspaces_example"}) // WorkspaceUser | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceUsersPost(context.Background()).WorkspaceUser(workspaceUser).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceUsersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1WorkspaceUsersPost`: WorkspaceUser
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1WorkspaceUsersPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WorkspaceUsersPostRequest struct via the builder pattern


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

