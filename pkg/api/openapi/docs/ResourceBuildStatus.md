# ResourceBuildStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**ImageUrl** | Pointer to **string** |  | [optional] 
**BuildSourceHash** | Pointer to **string** |  | [optional] 

## Methods

### NewResourceBuildStatus

`func NewResourceBuildStatus() *ResourceBuildStatus`

NewResourceBuildStatus instantiates a new ResourceBuildStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceBuildStatusWithDefaults

`func NewResourceBuildStatusWithDefaults() *ResourceBuildStatus`

NewResourceBuildStatusWithDefaults instantiates a new ResourceBuildStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *ResourceBuildStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ResourceBuildStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ResourceBuildStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ResourceBuildStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetConditions

`func (o *ResourceBuildStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ResourceBuildStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ResourceBuildStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ResourceBuildStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetImageUrl

`func (o *ResourceBuildStatus) GetImageUrl() string`

GetImageUrl returns the ImageUrl field if non-nil, zero value otherwise.

### GetImageUrlOk

`func (o *ResourceBuildStatus) GetImageUrlOk() (*string, bool)`

GetImageUrlOk returns a tuple with the ImageUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageUrl

`func (o *ResourceBuildStatus) SetImageUrl(v string)`

SetImageUrl sets ImageUrl field to given value.

### HasImageUrl

`func (o *ResourceBuildStatus) HasImageUrl() bool`

HasImageUrl returns a boolean if a field has been set.

### GetBuildSourceHash

`func (o *ResourceBuildStatus) GetBuildSourceHash() string`

GetBuildSourceHash returns the BuildSourceHash field if non-nil, zero value otherwise.

### GetBuildSourceHashOk

`func (o *ResourceBuildStatus) GetBuildSourceHashOk() (*string, bool)`

GetBuildSourceHashOk returns a tuple with the BuildSourceHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildSourceHash

`func (o *ResourceBuildStatus) SetBuildSourceHash(v string)`

SetBuildSourceHash sets BuildSourceHash field to given value.

### HasBuildSourceHash

`func (o *ResourceBuildStatus) HasBuildSourceHash() bool`

HasBuildSourceHash returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


