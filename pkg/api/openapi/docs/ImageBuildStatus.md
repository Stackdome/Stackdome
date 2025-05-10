# ImageBuildStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**ImageUrl** | Pointer to **string** |  | [optional] 
**BuildSourceRevision** | Pointer to **string** |  | [optional] 

## Methods

### NewImageBuildStatus

`func NewImageBuildStatus() *ImageBuildStatus`

NewImageBuildStatus instantiates a new ImageBuildStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImageBuildStatusWithDefaults

`func NewImageBuildStatusWithDefaults() *ImageBuildStatus`

NewImageBuildStatusWithDefaults instantiates a new ImageBuildStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *ImageBuildStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ImageBuildStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ImageBuildStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ImageBuildStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetConditions

`func (o *ImageBuildStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ImageBuildStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ImageBuildStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ImageBuildStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetImageUrl

`func (o *ImageBuildStatus) GetImageUrl() string`

GetImageUrl returns the ImageUrl field if non-nil, zero value otherwise.

### GetImageUrlOk

`func (o *ImageBuildStatus) GetImageUrlOk() (*string, bool)`

GetImageUrlOk returns a tuple with the ImageUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageUrl

`func (o *ImageBuildStatus) SetImageUrl(v string)`

SetImageUrl sets ImageUrl field to given value.

### HasImageUrl

`func (o *ImageBuildStatus) HasImageUrl() bool`

HasImageUrl returns a boolean if a field has been set.

### GetBuildSourceRevision

`func (o *ImageBuildStatus) GetBuildSourceRevision() string`

GetBuildSourceRevision returns the BuildSourceRevision field if non-nil, zero value otherwise.

### GetBuildSourceRevisionOk

`func (o *ImageBuildStatus) GetBuildSourceRevisionOk() (*string, bool)`

GetBuildSourceRevisionOk returns a tuple with the BuildSourceRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildSourceRevision

`func (o *ImageBuildStatus) SetBuildSourceRevision(v string)`

SetBuildSourceRevision sets BuildSourceRevision field to given value.

### HasBuildSourceRevision

`func (o *ImageBuildStatus) HasBuildSourceRevision() bool`

HasBuildSourceRevision returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


