# \ReleasesApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CancelRelease**](ReleasesApi.md#CancelRelease) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/releases/{release_id}/cancel | Cancel a pending or rendering release
[**CreateRelease**](ReleasesApi.md#CreateRelease) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/releases | Create a new release (deploy)
[**GetRelease**](ReleasesApi.md#GetRelease) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/releases/{release_id} | Get a release by ID
[**ListReleaseEvents**](ReleasesApi.md#ListReleaseEvents) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/releases/{release_id}/events | List release events ordered by sequence
[**ListReleases**](ReleasesApi.md#ListReleases) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/releases | List releases for a stack
[**StreamReleaseEvents**](ReleasesApi.md#StreamReleaseEvents) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/stacks/{id}/releases/{release_id}/events/stream | Stream release events via Server-Sent Events



## CancelRelease

> CancelRelease(ctx, orgId, projectName, id, releaseId).Execute()

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
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    releaseId := "releaseId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.CancelRelease(context.Background(), orgId, projectName, id, releaseId).Execute()
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
**projectName** | **string** | The name of the project | 
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

> StackRelease CreateRelease(ctx, orgId, projectName, id).CreateReleaseRequest(createReleaseRequest).Execute()

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
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    createReleaseRequest := *openapiclient.NewCreateReleaseRequest() // CreateReleaseRequest |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.CreateRelease(context.Background(), orgId, projectName, id).CreateReleaseRequest(createReleaseRequest).Execute()
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
**projectName** | **string** | The name of the project | 
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

> StackReleaseDetail GetRelease(ctx, orgId, projectName, id, releaseId).Execute()

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
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    releaseId := "releaseId_example" // string | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.GetRelease(context.Background(), orgId, projectName, id, releaseId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ReleasesApi.GetRelease``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetRelease`: StackReleaseDetail
    fmt.Fprintf(os.Stdout, "Response from `ReleasesApi.GetRelease`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**releaseId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetReleaseRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------





### Return type

[**StackReleaseDetail**](StackReleaseDetail.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListReleaseEvents

> ReleaseEventList ListReleaseEvents(ctx, orgId, projectName, id, releaseId).AfterSequence(afterSequence).Limit(limit).Execute()

List release events ordered by sequence

### Example

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
    releaseId := "releaseId_example" // string | 
    afterSequence := int32(56) // int32 |  (optional) (default to 0)
    limit := int32(56) // int32 |  (optional) (default to 100)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.ListReleaseEvents(context.Background(), orgId, projectName, id, releaseId).AfterSequence(afterSequence).Limit(limit).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ReleasesApi.ListReleaseEvents``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListReleaseEvents`: ReleaseEventList
    fmt.Fprintf(os.Stdout, "Response from `ReleasesApi.ListReleaseEvents`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**releaseId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiListReleaseEventsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **afterSequence** | **int32** |  | [default to 0]
 **limit** | **int32** |  | [default to 100]

### Return type

[**ReleaseEventList**](ReleaseEventList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListReleases

> StackReleaseList ListReleases(ctx, orgId, projectName, id).State(state).Page(page).PageSize(pageSize).Execute()

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
    projectName := "projectName_example" // string | The name of the project
    id := "id_example" // string | The id of record
    state := openapiclient.StackReleaseState("Pending") // StackReleaseState | Filter by release state (optional)
    page := int32(56) // int32 | Page number (optional) (default to 1)
    pageSize := int32(56) // int32 | Number of items per page (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.ListReleases(context.Background(), orgId, projectName, id).State(state).Page(page).PageSize(pageSize).Execute()
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
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 

### Other Parameters

Other parameters are passed through a pointer to a apiListReleasesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **state** | [**StackReleaseState**](StackReleaseState.md) | Filter by release state | 
 **page** | **int32** | Page number | [default to 1]
 **pageSize** | **int32** | Number of items per page | [default to 20]

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


## StreamReleaseEvents

> *os.File StreamReleaseEvents(ctx, orgId, projectName, id, releaseId).AfterSequence(afterSequence).Execute()

Stream release events via Server-Sent Events

### Example

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
    releaseId := "releaseId_example" // string | 
    afterSequence := int32(56) // int32 |  (optional) (default to 0)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.ReleasesApi.StreamReleaseEvents(context.Background(), orgId, projectName, id, releaseId).AfterSequence(afterSequence).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `ReleasesApi.StreamReleaseEvents``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `StreamReleaseEvents`: *os.File
    fmt.Fprintf(os.Stdout, "Response from `ReleasesApi.StreamReleaseEvents`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 
**id** | **string** | The id of record | 
**releaseId** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiStreamReleaseEventsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




 **afterSequence** | **int32** |  | [default to 0]

### Return type

[***os.File**](*os.File.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/event-stream

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

