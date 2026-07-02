# PreviewStackSync

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Commit** | Pointer to **string** |  | [optional] 
**StackfileContent** | Pointer to **string** |  | [optional] 
**ForceSync** | Pointer to **bool** |  | [optional] 
**ImageOverrides** | Pointer to **map[string]string** |  | [optional] 

## Methods

### NewPreviewStackSync

`func NewPreviewStackSync() *PreviewStackSync`

NewPreviewStackSync instantiates a new PreviewStackSync object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPreviewStackSyncWithDefaults

`func NewPreviewStackSyncWithDefaults() *PreviewStackSync`

NewPreviewStackSyncWithDefaults instantiates a new PreviewStackSync object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommit

`func (o *PreviewStackSync) GetCommit() string`

GetCommit returns the Commit field if non-nil, zero value otherwise.

### GetCommitOk

`func (o *PreviewStackSync) GetCommitOk() (*string, bool)`

GetCommitOk returns a tuple with the Commit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommit

`func (o *PreviewStackSync) SetCommit(v string)`

SetCommit sets Commit field to given value.

### HasCommit

`func (o *PreviewStackSync) HasCommit() bool`

HasCommit returns a boolean if a field has been set.

### GetStackfileContent

`func (o *PreviewStackSync) GetStackfileContent() string`

GetStackfileContent returns the StackfileContent field if non-nil, zero value otherwise.

### GetStackfileContentOk

`func (o *PreviewStackSync) GetStackfileContentOk() (*string, bool)`

GetStackfileContentOk returns a tuple with the StackfileContent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackfileContent

`func (o *PreviewStackSync) SetStackfileContent(v string)`

SetStackfileContent sets StackfileContent field to given value.

### HasStackfileContent

`func (o *PreviewStackSync) HasStackfileContent() bool`

HasStackfileContent returns a boolean if a field has been set.

### GetForceSync

`func (o *PreviewStackSync) GetForceSync() bool`

GetForceSync returns the ForceSync field if non-nil, zero value otherwise.

### GetForceSyncOk

`func (o *PreviewStackSync) GetForceSyncOk() (*bool, bool)`

GetForceSyncOk returns a tuple with the ForceSync field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForceSync

`func (o *PreviewStackSync) SetForceSync(v bool)`

SetForceSync sets ForceSync field to given value.

### HasForceSync

`func (o *PreviewStackSync) HasForceSync() bool`

HasForceSync returns a boolean if a field has been set.

### GetImageOverrides

`func (o *PreviewStackSync) GetImageOverrides() map[string]string`

GetImageOverrides returns the ImageOverrides field if non-nil, zero value otherwise.

### GetImageOverridesOk

`func (o *PreviewStackSync) GetImageOverridesOk() (*map[string]string, bool)`

GetImageOverridesOk returns a tuple with the ImageOverrides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageOverrides

`func (o *PreviewStackSync) SetImageOverrides(v map[string]string)`

SetImageOverrides sets ImageOverrides field to given value.

### HasImageOverrides

`func (o *PreviewStackSync) HasImageOverrides() bool`

HasImageOverrides returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


