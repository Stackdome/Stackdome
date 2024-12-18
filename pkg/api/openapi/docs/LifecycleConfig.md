# LifecycleConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RestartRequestTime** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewLifecycleConfig

`func NewLifecycleConfig() *LifecycleConfig`

NewLifecycleConfig instantiates a new LifecycleConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLifecycleConfigWithDefaults

`func NewLifecycleConfigWithDefaults() *LifecycleConfig`

NewLifecycleConfigWithDefaults instantiates a new LifecycleConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRestartRequestTime

`func (o *LifecycleConfig) GetRestartRequestTime() time.Time`

GetRestartRequestTime returns the RestartRequestTime field if non-nil, zero value otherwise.

### GetRestartRequestTimeOk

`func (o *LifecycleConfig) GetRestartRequestTimeOk() (*time.Time, bool)`

GetRestartRequestTimeOk returns a tuple with the RestartRequestTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartRequestTime

`func (o *LifecycleConfig) SetRestartRequestTime(v time.Time)`

SetRestartRequestTime sets RestartRequestTime field to given value.

### HasRestartRequestTime

`func (o *LifecycleConfig) HasRestartRequestTime() bool`

HasRestartRequestTime returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


