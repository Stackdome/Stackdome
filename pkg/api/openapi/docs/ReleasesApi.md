# \ReleasesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CancelRelease**](ReleasesApi.md#CancelRelease) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/releases/{release_id}/cancel | Cancel a pending or rendering release
[**CreateRelease**](ReleasesApi.md#CreateRelease) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/releases | Create a new release (deploy)
[**GetRelease**](ReleasesApi.md#GetRelease) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/releases/{release_id} | Get a release by ID
[**ListReleases**](ReleasesApi.md#ListReleases) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}/releases | List releases for a stack



## CancelRelease

> CancelRelease(ctx, orgId, teamName, id, releaseId).Execute()

Cancel a pending or rendering release

### Example

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
    releaseId := "releaseId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.CancelRelease(context.Background(), orgId, teamName, id, releaseId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ReleasesApi.CancelRelease``: %v\n", err)
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
**releaseId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiCancelReleaseRequest struct via the builder pattern


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


## CreateRelease

> StackRelease CreateRelease(ctx, orgId, teamName, id).CreateReleaseRequest(createReleaseRequest).Execute()

Create a new release (deploy)

### Example

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
    createReleaseRequest := *openapiclient.NewCreateReleaseRequest() // CreateReleaseRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.CreateRelease(context.Background(), orgId, teamName, id).CreateReleaseRequest(createReleaseRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ReleasesApi.CreateRelease``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreateRelease`: StackRelease
    fmt.Fprintf(os.Stdout, "Response from `ReleasesApi.CreateRelease`: %v\n", resp)
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

Other parameters are passed through a pointer to a apiCreateReleaseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **createReleaseRequest** | [**CreateReleaseRequest**](CreateReleaseRequest.md) |  | 

### Return type

[**StackRelease**](StackRelease.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetRelease

> StackRelease GetRelease(ctx, orgId, teamName, id, releaseId).Execute()

Get a release by ID

### Example

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
    releaseId := "releaseId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.GetRelease(context.Background(), orgId, teamName, id, releaseId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ReleasesApi.GetRelease``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRelease`: StackRelease
    fmt.Fprintf(os.Stdout, "Response from `ReleasesApi.GetRelease`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 
**id** | **string** | The id of record | 
**releaseId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetReleaseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**StackRelease**](StackRelease.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListReleases

> StackReleaseList ListReleases(ctx, orgId, teamName, id).Execute()

List releases for a stack

### Example

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
    resp, r, err := apiClient.ReleasesApi.ListReleases(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ReleasesApi.ListReleases``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListReleases`: StackReleaseList
    fmt.Fprintf(os.Stdout, "Response from `ReleasesApi.ListReleases`: %v\n", resp)
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

Other parameters are passed through a pointer to a apiListReleasesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**StackReleaseList**](StackReleaseList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

