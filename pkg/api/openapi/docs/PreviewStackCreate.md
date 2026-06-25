# PreviewStackCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigId** | **string** |  | 
**PrNumber** | **string** |  | 
**Branch** | **string** |  | 
**Commit** | Pointer to **string** |  | [optional] 
**StackfileContent** | Pointer to **string** |  | [optional] 
**ImageOverrides** | Pointer to **map[string]string** |  | [optional] 

## Methods

### NewPreviewStackCreate

`func NewPreviewStackCreate(configId string, prNumber string, branch string, ) *PreviewStackCreate`

NewPreviewStackCreate instantiates a new PreviewStackCreate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPreviewStackCreateWithDefaults

`func NewPreviewStackCreateWithDefaults() *PreviewStackCreate`

NewPreviewStackCreateWithDefaults instantiates a new PreviewStackCreate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigId

`func (o *PreviewStackCreate) GetConfigId() string`

GetConfigId returns the ConfigId field if non-nil, zero value otherwise.

### GetConfigIdOk

`func (o *PreviewStackCreate) GetConfigIdOk() (*string, bool)`

GetConfigIdOk returns a tuple with the ConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigId

`func (o *PreviewStackCreate) SetConfigId(v string)`

SetConfigId sets ConfigId field to given value.


### GetPrNumber

`func (o *PreviewStackCreate) GetPrNumber() string`

GetPrNumber returns the PrNumber field if non-nil, zero value otherwise.

### GetPrNumberOk

`func (o *PreviewStackCreate) GetPrNumberOk() (*string, bool)`

GetPrNumberOk returns a tuple with the PrNumber field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrNumber

`func (o *PreviewStackCreate) SetPrNumber(v string)`

SetPrNumber sets PrNumber field to given value.


### GetBranch

`func (o *PreviewStackCreate) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *PreviewStackCreate) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *PreviewStackCreate) SetBranch(v string)`

SetBranch sets Branch field to given value.


### GetCommit

`func (o *PreviewStackCreate) GetCommit() string`

GetCommit returns the Commit field if non-nil, zero value otherwise.

### GetCommitOk

`func (o *PreviewStackCreate) GetCommitOk() (*string, bool)`

GetCommitOk returns a tuple with the Commit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommit

`func (o *PreviewStackCreate) SetCommit(v string)`

SetCommit sets Commit field to given value.

### HasCommit

`func (o *PreviewStackCreate) HasCommit() bool`

HasCommit returns a boolean if a field has been set.

### GetStackfileContent

`func (o *PreviewStackCreate) GetStackfileContent() string`

GetStackfileContent returns the StackfileContent field if non-nil, zero value otherwise.

### GetStackfileContentOk

`func (o *PreviewStackCreate) GetStackfileContentOk() (*string, bool)`

GetStackfileContentOk returns a tuple with the StackfileContent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackfileContent

`func (o *PreviewStackCreate) SetStackfileContent(v string)`

SetStackfileContent sets StackfileContent field to given value.

### HasStackfileContent

`func (o *PreviewStackCreate) HasStackfileContent() bool`

HasStackfileContent returns a boolean if a field has been set.

### GetImageOverrides

`func (o *PreviewStackCreate) GetImageOverrides() map[string]string`

GetImageOverrides returns the ImageOverrides field if non-nil, zero value otherwise.

### GetImageOverridesOk

`func (o *PreviewStackCreate) GetImageOverridesOk() (*map[string]string, bool)`

GetImageOverridesOk returns a tuple with the ImageOverrides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageOverrides

`func (o *PreviewStackCreate) SetImageOverrides(v map[string]string)`

SetImageOverrides sets ImageOverrides field to given value.

### HasImageOverrides

`func (o *PreviewStackCreate) HasImageOverrides() bool`

HasImageOverrides returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


