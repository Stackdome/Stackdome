# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ApiV1AuthLoginPost**](DefaultApi.md#ApiV1AuthLoginPost) | **Post** /api/v1/auth/login | User login
[**ApiV1OrganizationsIdWorkspaceStoragesGet**](DefaultApi.md#ApiV1OrganizationsIdWorkspaceStoragesGet) | **Get** /api/v1/organizations/{id}/workspace-storages | List all WorkspaceStorage objects for an organization
[**ApiV1OrganizationsIdWorkspaceStoragesPost**](DefaultApi.md#ApiV1OrganizationsIdWorkspaceStoragesPost) | **Post** /api/v1/organizations/{id}/workspace-storages | Create a new WorkspaceStorage object
[**ApiV1OrganizationsOrgIdWorkspaceStoragesIdDelete**](DefaultApi.md#ApiV1OrganizationsOrgIdWorkspaceStoragesIdDelete) | **Delete** /api/v1/organizations/{org_id}/workspace-storages/{id} | Delete a WorkspaceStorage object
[**ApiV1OrganizationsOrgIdWorkspaceStoragesIdGet**](DefaultApi.md#ApiV1OrganizationsOrgIdWorkspaceStoragesIdGet) | **Get** /api/v1/organizations/{org_id}/workspace-storages/{id} | Get a specific WorkspaceStorage object
[**ApiV1OrganizationsOrgIdWorkspaceStoragesIdPut**](DefaultApi.md#ApiV1OrganizationsOrgIdWorkspaceStoragesIdPut) | **Put** /api/v1/organizations/{org_id}/workspace-storages/{id} | Update a WorkspaceStorage object
[**ApiV1UsersIdGet**](DefaultApi.md#ApiV1UsersIdGet) | **Get** /api/v1/users/{id} | Get a user
[**ApiV1UsersMeGet**](DefaultApi.md#ApiV1UsersMeGet) | **Get** /api/v1/users/me | Get a the current authenticated user
[**ApiV1UsersPost**](DefaultApi.md#ApiV1UsersPost) | **Post** /api/v1/users | Create new user
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


## ApiV1OrganizationsIdWorkspaceStoragesGet

> ApiV1OrganizationsIdWorkspaceStoragesGet200Response ApiV1OrganizationsIdWorkspaceStoragesGet(ctx, id).Limit(limit).Offset(offset).WorkspaceName(workspaceName).State(state).Execute()

List all WorkspaceStorage objects for an organization

### Example

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
    limit := int32(56) // int32 |  (optional) (default to 20)
    offset := int32(56) // int32 |  (optional) (default to 0)
    workspaceName := "workspaceName_example" // string |  (optional)
    state := "state_example" // string |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsIdWorkspaceStoragesGet(context.Background(), id).Limit(limit).Offset(offset).WorkspaceName(workspaceName).State(state).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsIdWorkspaceStoragesGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsIdWorkspaceStoragesGet`: ApiV1OrganizationsIdWorkspaceStoragesGet200Response
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsIdWorkspaceStoragesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsIdWorkspaceStoragesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **limit** | **int32** |  | [default to 20]
 **offset** | **int32** |  | [default to 0]
 **workspaceName** | **string** |  | 
 **state** | **string** |  | 

### Return type

[**ApiV1OrganizationsIdWorkspaceStoragesGet200Response**](ApiV1OrganizationsIdWorkspaceStoragesGet200Response.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsIdWorkspaceStoragesPost

> WorkspaceStorage ApiV1OrganizationsIdWorkspaceStoragesPost(ctx, id).WorkspaceStorage(workspaceStorage).Execute()

Create a new WorkspaceStorage object

### Example

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
    workspaceStorage := *openapiclient.NewWorkspaceStorage("Name_example", *openapiclient.NewWorkspaceStorageSpec("WorkspaceName_example")) // WorkspaceStorage | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsIdWorkspaceStoragesPost(context.Background(), id).WorkspaceStorage(workspaceStorage).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsIdWorkspaceStoragesPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsIdWorkspaceStoragesPost`: WorkspaceStorage
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsIdWorkspaceStoragesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsIdWorkspaceStoragesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workspaceStorage** | [**WorkspaceStorage**](WorkspaceStorage.md) |  | 

### Return type

[**WorkspaceStorage**](WorkspaceStorage.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdWorkspaceStoragesIdDelete

> ApiV1OrganizationsOrgIdWorkspaceStoragesIdDelete(ctx, orgId, id).Execute()

Delete a WorkspaceStorage object

### Example

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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdWorkspaceStoragesIdDelete(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdWorkspaceStoragesIdDelete``: %v\n", err)
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

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdWorkspaceStoragesIdDeleteRequest struct via the builder pattern


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


## ApiV1OrganizationsOrgIdWorkspaceStoragesIdGet

> WorkspaceStorage ApiV1OrganizationsOrgIdWorkspaceStoragesIdGet(ctx, orgId, id).Execute()

Get a specific WorkspaceStorage object

### Example

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
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdWorkspaceStoragesIdGet(context.Background(), orgId, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdWorkspaceStoragesIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdWorkspaceStoragesIdGet`: WorkspaceStorage
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdWorkspaceStoragesIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdWorkspaceStoragesIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**WorkspaceStorage**](WorkspaceStorage.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1OrganizationsOrgIdWorkspaceStoragesIdPut

> WorkspaceStorage ApiV1OrganizationsOrgIdWorkspaceStoragesIdPut(ctx, orgId, id).WorkspaceStorage(workspaceStorage).Execute()

Update a WorkspaceStorage object

### Example

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
    workspaceStorage := *openapiclient.NewWorkspaceStorage("Name_example", *openapiclient.NewWorkspaceStorageSpec("WorkspaceName_example")) // WorkspaceStorage | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdWorkspaceStoragesIdPut(context.Background(), orgId, id).WorkspaceStorage(workspaceStorage).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1OrganizationsOrgIdWorkspaceStoragesIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1OrganizationsOrgIdWorkspaceStoragesIdPut`: WorkspaceStorage
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1OrganizationsOrgIdWorkspaceStoragesIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** |  | 
**id** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1OrganizationsOrgIdWorkspaceStoragesIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **workspaceStorage** | [**WorkspaceStorage**](WorkspaceStorage.md) |  | 

### Return type

[**WorkspaceStorage**](WorkspaceStorage.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
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


## ApiV1UsersMeGet

> User ApiV1UsersMeGet(ctx).Execute()

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
    resp, r, err := apiClient.DefaultApi.ApiV1UsersMeGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1UsersMeGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1UsersMeGet`: User
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1UsersMeGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1UsersMeGetRequest struct via the builder pattern


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


## ApiV1UsersPost

> User ApiV1UsersPost(ctx).UserCreateRequest(userCreateRequest).Execute()

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
    userCreateRequest := *openapiclient.NewUserCreateRequest("Name_example", "Email_example", "Password_example", "OrganisationId_example") // UserCreateRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1UsersPost(context.Background()).UserCreateRequest(userCreateRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1UsersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1UsersPost`: User
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1UsersPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1UsersPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userCreateRequest** | [**UserCreateRequest**](UserCreateRequest.md) |  | 

### Return type

[**User**](User.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
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
    workspaceUser := *openapiclient.NewWorkspaceUser("SshPublicKey_example", []string{"Workspaces_example"}) // WorkspaceUser | 

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
    workspaceUser := *openapiclient.NewWorkspaceUser("SshPublicKey_example", []string{"Workspaces_example"}) // WorkspaceUser | 

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

