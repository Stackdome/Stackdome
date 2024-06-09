# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**ApiV1AuthLoginPost**](DefaultApi.md#ApiV1AuthLoginPost) | **Post** /api/v1/auth/login | User login
[**ApiV1ClustersGet**](DefaultApi.md#ApiV1ClustersGet) | **Get** /api/v1/clusters | Get all clusters
[**ApiV1ClustersIdDelete**](DefaultApi.md#ApiV1ClustersIdDelete) | **Delete** /api/v1/clusters/{id} | Delete a cluster
[**ApiV1ClustersIdGet**](DefaultApi.md#ApiV1ClustersIdGet) | **Get** /api/v1/clusters/{id} | Get a cluster by ID
[**ApiV1ClustersIdPut**](DefaultApi.md#ApiV1ClustersIdPut) | **Put** /api/v1/clusters/{id} | Update a cluster
[**ApiV1ClustersPost**](DefaultApi.md#ApiV1ClustersPost) | **Post** /api/v1/clusters | Create a new cluster
[**ApiV1UsersIdGet**](DefaultApi.md#ApiV1UsersIdGet) | **Get** /api/v1/users/{id} | Get a user
[**ApiV1UsersPost**](DefaultApi.md#ApiV1UsersPost) | **Post** /api/v1/users | Create new user
[**ApiV1WorkspaceProvisionRequestsIdDelete**](DefaultApi.md#ApiV1WorkspaceProvisionRequestsIdDelete) | **Delete** /api/v1/workspace-provision-requests/{id} | Delete a WorkspaceProvisionRequest
[**ApiV1WorkspaceProvisionRequestsIdGet**](DefaultApi.md#ApiV1WorkspaceProvisionRequestsIdGet) | **Get** /api/v1/workspace-provision-requests/{id} | Get a workspace provision request object by ID
[**ApiV1WorkspaceProvisionRequestsIdPut**](DefaultApi.md#ApiV1WorkspaceProvisionRequestsIdPut) | **Put** /api/v1/workspace-provision-requests/{id} | Update a WorkspaceProvisionRequest
[**ApiV1WorkspaceProvisionRequestsPost**](DefaultApi.md#ApiV1WorkspaceProvisionRequestsPost) | **Post** /api/v1/workspace-provision-requests | Create a new workpace provision request object



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


## ApiV1ClustersGet

> []Cluster ApiV1ClustersGet(ctx).Execute()

Get all clusters

### Example

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
    resp, r, err := apiClient.DefaultApi.ApiV1ClustersGet(context.Background()).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ClustersGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ClustersGet`: []Cluster
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ClustersGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ClustersGetRequest struct via the builder pattern


### Return type

[**[]Cluster**](Cluster.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ClustersIdDelete

> ApiV1ClustersIdDelete(ctx, id).Execute()

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
    id := "id_example" // string | The id of record

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ClustersIdDelete(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ClustersIdDelete``: %v\n", err)
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

Other parameters are passed through a pointer to a apiApiV1ClustersIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## ApiV1ClustersIdGet

> Cluster ApiV1ClustersIdGet(ctx, id).Execute()

Get a cluster by ID

### Example

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
    resp, r, err := apiClient.DefaultApi.ApiV1ClustersIdGet(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ClustersIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ClustersIdGet`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ClustersIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ClustersIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**Cluster**](Cluster.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ClustersIdPut

> Cluster ApiV1ClustersIdPut(ctx, id).Cluster(cluster).Execute()

Update a cluster

### Example

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
    cluster := *openapiclient.NewCluster() // Cluster | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ClustersIdPut(context.Background(), id).Cluster(cluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ClustersIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ClustersIdPut`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ClustersIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ClustersIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **cluster** | [**Cluster**](Cluster.md) |  | 

### Return type

[**Cluster**](Cluster.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1ClustersPost

> Cluster ApiV1ClustersPost(ctx).Cluster(cluster).Execute()

Create a new cluster

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    cluster := *openapiclient.NewCluster() // Cluster | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1ClustersPost(context.Background()).Cluster(cluster).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1ClustersPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1ClustersPost`: Cluster
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1ClustersPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1ClustersPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **cluster** | [**Cluster**](Cluster.md) |  | 

### Return type

[**Cluster**](Cluster.md)

### Authorization

No authorization required

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

No authorization required

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
    userCreateRequest := *openapiclient.NewUserCreateRequest("Name_example", "Email_example", "Password_example", int32(123)) // UserCreateRequest | 

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


## ApiV1WorkspaceProvisionRequestsIdDelete

> ApiV1WorkspaceProvisionRequestsIdDelete(ctx, id).Execute()

Delete a WorkspaceProvisionRequest

### Example

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
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceProvisionRequestsIdDelete(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceProvisionRequestsIdDelete``: %v\n", err)
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

Other parameters are passed through a pointer to a apiApiV1WorkspaceProvisionRequestsIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


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


## ApiV1WorkspaceProvisionRequestsIdGet

> WorkspaceProvisionRequest ApiV1WorkspaceProvisionRequestsIdGet(ctx, id).Execute()

Get a workspace provision request object by ID

### Example

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
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceProvisionRequestsIdGet(context.Background(), id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceProvisionRequestsIdGet``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1WorkspaceProvisionRequestsIdGet`: WorkspaceProvisionRequest
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1WorkspaceProvisionRequestsIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WorkspaceProvisionRequestsIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**WorkspaceProvisionRequest**](WorkspaceProvisionRequest.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1WorkspaceProvisionRequestsIdPut

> WorkspaceProvisionRequest ApiV1WorkspaceProvisionRequestsIdPut(ctx, id).WorkspaceProvisionRequest(workspaceProvisionRequest).Execute()

Update a WorkspaceProvisionRequest

### Example

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
    workspaceProvisionRequest := *openapiclient.NewWorkspaceProvisionRequest("SshPublicKey_example") // WorkspaceProvisionRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceProvisionRequestsIdPut(context.Background(), id).WorkspaceProvisionRequest(workspaceProvisionRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceProvisionRequestsIdPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1WorkspaceProvisionRequestsIdPut`: WorkspaceProvisionRequest
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1WorkspaceProvisionRequestsIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WorkspaceProvisionRequestsIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **workspaceProvisionRequest** | [**WorkspaceProvisionRequest**](WorkspaceProvisionRequest.md) |  | 

### Return type

[**WorkspaceProvisionRequest**](WorkspaceProvisionRequest.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ApiV1WorkspaceProvisionRequestsPost

> WorkspaceProvisionRequest ApiV1WorkspaceProvisionRequestsPost(ctx).WorkspaceProvisionRequest(workspaceProvisionRequest).Execute()

Create a new workpace provision request object

### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    workspaceProvisionRequest := *openapiclient.NewWorkspaceProvisionRequest("SshPublicKey_example") // WorkspaceProvisionRequest | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.DefaultApi.ApiV1WorkspaceProvisionRequestsPost(context.Background()).WorkspaceProvisionRequest(workspaceProvisionRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.ApiV1WorkspaceProvisionRequestsPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ApiV1WorkspaceProvisionRequestsPost`: WorkspaceProvisionRequest
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.ApiV1WorkspaceProvisionRequestsPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiApiV1WorkspaceProvisionRequestsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workspaceProvisionRequest** | [**WorkspaceProvisionRequest**](WorkspaceProvisionRequest.md) |  | 

### Return type

[**WorkspaceProvisionRequest**](WorkspaceProvisionRequest.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

