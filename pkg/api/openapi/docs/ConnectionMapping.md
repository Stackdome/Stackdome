# ConnectionMapping

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Target** | [**ConnectionTarget**](ConnectionTarget.md) |  | 
**Value** | [**ValueRef**](ValueRef.md) |  | 

## Methods

### NewConnectionMapping

`func NewConnectionMapping(target ConnectionTarget, value ValueRef, ) *ConnectionMapping`

NewConnectionMapping instantiates a new ConnectionMapping object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectionMappingWithDefaults

`func NewConnectionMappingWithDefaults() *ConnectionMapping`

NewConnectionMappingWithDefaults instantiates a new ConnectionMapping object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTarget

`func (o *ConnectionMapping) GetTarget() ConnectionTarget`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *ConnectionMapping) GetTargetOk() (*ConnectionTarget, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *ConnectionMapping) SetTarget(v ConnectionTarget)`

SetTarget sets Target field to given value.


### GetValue

`func (o *ConnectionMapping) GetValue() ValueRef`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *ConnectionMapping) GetValueOk() (*ValueRef, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *ConnectionMapping) SetValue(v ValueRef)`

SetValue sets Value field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


