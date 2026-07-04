# \PreviewConfigsApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreatePreviewConfig**](PreviewConfigsApi.md#CreatePreviewConfig) | **Post** /api/v1/organizations/{org_id}/teams/{team_name}/stack-preview-configs | Create a new preview config
[**DeletePreviewConfig**](PreviewConfigsApi.md#DeletePreviewConfig) | **Delete** /api/v1/organizations/{org_id}/teams/{team_name}/stack-preview-configs/{id} | Delete a preview config
[**GetPreviewConfig**](PreviewConfigsApi.md#GetPreviewConfig) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stack-preview-configs/{id} | Get a specific preview config
[**ListPreviewConfigs**](PreviewConfigsApi.md#ListPreviewConfigs) | **Get** /api/v1/organizations/{org_id}/teams/{team_name}/stack-preview-configs | List preview configs for a team
[**UpdatePreviewConfig**](PreviewConfigsApi.md#UpdatePreviewConfig) | **Put** /api/v1/organizations/{org_id}/teams/{team_name}/stack-preview-configs/{id} | Update a preview config



## CreatePreviewConfig

> StackPreviewConfig CreatePreviewConfig(ctx, orgId, teamName).StackPreviewConfigCreate(stackPreviewConfigCreate).Execute()

Create a new preview config

### Example

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
    stackPreviewConfigCreate := *openapiclient.NewStackPreviewConfigCreate("Name_example", *openapiclient.NewPreviewGitRepository("RepoUrl_example")) // StackPreviewConfigCreate | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PreviewConfigsApi.CreatePreviewConfig(context.Background(), orgId, teamName).StackPreviewConfigCreate(stackPreviewConfigCreate).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewConfigsApi.CreatePreviewConfig``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `CreatePreviewConfig`: StackPreviewConfig
    fmt.Fprintf(os.Stdout, "Response from `PreviewConfigsApi.CreatePreviewConfig`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiCreatePreviewConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **stackPreviewConfigCreate** | [**StackPreviewConfigCreate**](StackPreviewConfigCreate.md) |  | 

### Return type

[**StackPreviewConfig**](StackPreviewConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeletePreviewConfig

> DeletePreviewConfig(ctx, orgId, teamName, id).Execute()

Delete a preview config

### Example

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
    resp, r, err := apiClient.PreviewConfigsApi.DeletePreviewConfig(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewConfigsApi.DeletePreviewConfig``: %v\n", err)
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

Other parameters are passed through a pointer to a apiDeletePreviewConfigRequest struct via the builder pattern


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


## GetPreviewConfig

> StackPreviewConfig GetPreviewConfig(ctx, orgId, teamName, id).Execute()

Get a specific preview config

### Example

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
    resp, r, err := apiClient.PreviewConfigsApi.GetPreviewConfig(context.Background(), orgId, teamName, id).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewConfigsApi.GetPreviewConfig``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetPreviewConfig`: StackPreviewConfig
    fmt.Fprintf(os.Stdout, "Response from `PreviewConfigsApi.GetPreviewConfig`: %v\n", resp)
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

Other parameters are passed through a pointer to a apiGetPreviewConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**StackPreviewConfig**](StackPreviewConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## ListPreviewConfigs

> StackPreviewConfigList ListPreviewConfigs(ctx, orgId, teamName).Page(page).PageSize(pageSize).Execute()

List preview configs for a team

### Example

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
    page := int32(56) // int32 | Page number (optional) (default to 1)
    pageSize := int32(56) // int32 | Number of items per page (optional) (default to 20)

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PreviewConfigsApi.ListPreviewConfigs(context.Background(), orgId, teamName).Page(page).PageSize(pageSize).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewConfigsApi.ListPreviewConfigs``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `ListPreviewConfigs`: StackPreviewConfigList
    fmt.Fprintf(os.Stdout, "Response from `PreviewConfigsApi.ListPreviewConfigs`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**orgId** | **string** | The ID of the organization | 
**teamName** | **string** | The name of the team | 

### Other Parameters

Other parameters are passed through a pointer to a apiListPreviewConfigsRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **page** | **int32** | Page number | [default to 1]
 **pageSize** | **int32** | Number of items per page | [default to 20]

### Return type

[**StackPreviewConfigList**](StackPreviewConfigList.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UpdatePreviewConfig

> StackPreviewConfig UpdatePreviewConfig(ctx, orgId, teamName, id).StackPreviewConfigUpdate(stackPreviewConfigUpdate).Execute()

Update a preview config

### Example

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
    stackPreviewConfigUpdate := *openapiclient.NewStackPreviewConfigUpdate() // StackPreviewConfigUpdate | 

    configuration := openapiclient.NewConfiguration()
    apiClient := openapiclient.NewAPIClient(configuration)
    resp, r, err := apiClient.PreviewConfigsApi.UpdatePreviewConfig(context.Background(), orgId, teamName, id).StackPreviewConfigUpdate(stackPreviewConfigUpdate).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `PreviewConfigsApi.UpdatePreviewConfig``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `UpdatePreviewConfig`: StackPreviewConfig
    fmt.Fprintf(os.Stdout, "Response from `PreviewConfigsApi.UpdatePreviewConfig`: %v\n", resp)
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

Other parameters are passed through a pointer to a apiUpdatePreviewConfigRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **stackPreviewConfigUpdate** | [**StackPreviewConfigUpdate**](StackPreviewConfigUpdate.md) |  | 

### Return type

[**StackPreviewConfig**](StackPreviewConfig.md)

### Authorization

[Bearer](../README.md#Bearer)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

