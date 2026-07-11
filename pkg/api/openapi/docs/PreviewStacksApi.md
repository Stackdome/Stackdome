# \PreviewStacksApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePreviewStack**](PreviewStacksApi.md#CreatePreviewStack) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/preview-stacks | Create a new preview stack
[**DeletePreviewStack**](PreviewStacksApi.md#DeletePreviewStack) | **Delete** /api/v1/organizations/{org_id}/projects/{project_name}/preview-stacks/{id} | Delete a preview stack
[**GetPreviewStack**](PreviewStacksApi.md#GetPreviewStack) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/preview-stacks/{id} | Get a specific preview stack
[**ListPreviewStacks**](PreviewStacksApi.md#ListPreviewStacks) | **Get** /api/v1/organizations/{org_id}/projects/{project_name}/preview-stacks | List preview stacks for a project
[**SyncPreviewStack**](PreviewStacksApi.md#SyncPreviewStack) | **Post** /api/v1/organizations/{org_id}/projects/{project_name}/preview-stacks/{id}/sync | Sync a preview stack



## CreatePreviewStack

> PreviewStack CreatePreviewStack(ctx, orgId, projectName).PreviewStackCreate(previewStackCreate).Execute()

Create a new preview stack

### Example

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
    previewStackCreate := *openapiclient.NewPreviewStackCreate("ConfigId_example", "PrNumber_example", "Branch_example") // PreviewStackCreate | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PreviewStacksApi.CreatePreviewStack(context.Background(), orgId, projectName).PreviewStackCreate(previewStackCreate).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewStacksApi.CreatePreviewStack``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreatePreviewStack`: PreviewStack
    fmt.Fprintf(os.Stdout, "Response from `PreviewStacksApi.CreatePreviewStack`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreatePreviewStackRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **previewStackCreate** | [**PreviewStackCreate**](PreviewStackCreate.md) |  | 

### Return type

[**PreviewStack**](PreviewStack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePreviewStack

> PreviewStack DeletePreviewStack(ctx, orgId, projectName, id).Execute()

Delete a preview stack

### Example

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
    resp, r, err := apiClient.PreviewStacksApi.DeletePreviewStack(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewStacksApi.DeletePreviewStack``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `DeletePreviewStack`: PreviewStack
    fmt.Fprintf(os.Stdout, "Response from `PreviewStacksApi.DeletePreviewStack`: %v\n", resp)
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

Other parameters are passed through a pointer to a apiDeletePreviewStackRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**PreviewStack**](PreviewStack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetPreviewStack

> PreviewStack GetPreviewStack(ctx, orgId, projectName, id).Execute()

Get a specific preview stack

### Example

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
    resp, r, err := apiClient.PreviewStacksApi.GetPreviewStack(context.Background(), orgId, projectName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewStacksApi.GetPreviewStack``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPreviewStack`: PreviewStack
    fmt.Fprintf(os.Stdout, "Response from `PreviewStacksApi.GetPreviewStack`: %v\n", resp)
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

Other parameters are passed through a pointer to a apiGetPreviewStackRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**PreviewStack**](PreviewStack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPreviewStacks

> PreviewStackList ListPreviewStacks(ctx, orgId, projectName).Page(page).PageSize(pageSize).ConfigId(configId).Execute()

List preview stacks for a project

### Example

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
    page := int32(56) // int32 | Page number (optional) (default to 1)
    pageSize := int32(56) // int32 | Number of items per page (optional) (default to 20)
    configId := "configId_example" // string | Filter by preview config ID (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PreviewStacksApi.ListPreviewStacks(context.Background(), orgId, projectName).Page(page).PageSize(pageSize).ConfigId(configId).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewStacksApi.ListPreviewStacks``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListPreviewStacks`: PreviewStackList
    fmt.Fprintf(os.Stdout, "Response from `PreviewStacksApi.ListPreviewStacks`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**projectName** | **string** | The name of the project | 

### Other Parameters

Other parameters are passed through a pointer to a apiListPreviewStacksRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **page** | **int32** | Page number | [default to 1]
 **pageSize** | **int32** | Number of items per page | [default to 20]
 **configId** | **string** | Filter by preview config ID | 

### Return type

[**PreviewStackList**](PreviewStackList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SyncPreviewStack

> PreviewStack SyncPreviewStack(ctx, orgId, projectName, id).PreviewStackSync(previewStackSync).Execute()

Sync a preview stack

### Example

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
    previewStackSync := *openapiclient.NewPreviewStackSync() // PreviewStackSync |  (optional)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PreviewStacksApi.SyncPreviewStack(context.Background(), orgId, projectName, id).PreviewStackSync(previewStackSync).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewStacksApi.SyncPreviewStack``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `SyncPreviewStack`: PreviewStack
    fmt.Fprintf(os.Stdout, "Response from `PreviewStacksApi.SyncPreviewStack`: %v\n", resp)
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

Other parameters are passed through a pointer to a apiSyncPreviewStackRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **previewStackSync** | [**PreviewStackSync**](PreviewStackSync.md) |  | 

### Return type

[**PreviewStack**](PreviewStack.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

