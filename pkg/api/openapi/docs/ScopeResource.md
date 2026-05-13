# ScopeResource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resource** | Pointer to **string** | The resource name (e.g., stacks, secrets, addons/postgres) | [optional] 
**Actions** | Pointer to **[]string** | The allowed actions for this resource (e.g., read, write, delete) | [optional] 

## Methods

### NewScopeResource

`func NewScopeResource() *ScopeResource`

NewScopeResource instantiates a new ScopeResource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewScopeResourceWithDefaults

`func NewScopeResourceWithDefaults() *ScopeResource`

NewScopeResourceWithDefaults instantiates a new ScopeResource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResource

`func (o *ScopeResource) GetResource() string`

GetResource returns the Resource field if non-nil, zero value otherwise.

### GetResourceOk

`func (o *ScopeResource) GetResourceOk() (*string, bool)`

GetResourceOk returns a tuple with the Resource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResource

`func (o *ScopeResource) SetResource(v string)`

SetResource sets Resource field to given value.

### HasResource

`func (o *ScopeResource) HasResource() bool`

HasResource returns a boolean if a field has been set.

### GetActions

`func (o *ScopeResource) GetActions() []string`

GetActions returns the Actions field if non-nil, zero value otherwise.

### GetActionsOk

`func (o *ScopeResource) GetActionsOk() (*[]string, bool)`

GetActionsOk returns a tuple with the Actions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetActions

`func (o *ScopeResource) SetActions(v []string)`

SetActions sets Actions field to given value.

### HasActions

`func (o *ScopeResource) HasActions() bool`

HasActions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


