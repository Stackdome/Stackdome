# StackSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StackResources** | [**[]StackResource**](StackResource.md) |  | 

## Methods

### NewStackSpec

`func NewStackSpec(stackResources []StackResource, ) *StackSpec`

NewStackSpec instantiates a new StackSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackSpecWithDefaults

`func NewStackSpecWithDefaults() *StackSpec`

NewStackSpecWithDefaults instantiates a new StackSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStackResources

`func (o *StackSpec) GetStackResources() []StackResource`

GetStackResources returns the StackResources field if non-nil, zero value otherwise.

### GetStackResourcesOk

`func (o *StackSpec) GetStackResourcesOk() (*[]StackResource, bool)`

GetStackResourcesOk returns a tuple with the StackResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackResources

`func (o *StackSpec) SetStackResources(v []StackResource)`

SetStackResources sets StackResources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


