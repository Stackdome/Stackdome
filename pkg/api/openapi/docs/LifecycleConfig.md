# LifecycleConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**LastRestartRequestTime** | Pointer to **time.Time** |  | [optional] 

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

### GetLastRestartRequestTime

`func (o *LifecycleConfig) GetLastRestartRequestTime() time.Time`

GetLastRestartRequestTime returns the LastRestartRequestTime field if non-nil, zero value otherwise.

### GetLastRestartRequestTimeOk

`func (o *LifecycleConfig) GetLastRestartRequestTimeOk() (*time.Time, bool)`

GetLastRestartRequestTimeOk returns a tuple with the LastRestartRequestTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastRestartRequestTime

`func (o *LifecycleConfig) SetLastRestartRequestTime(v time.Time)`

SetLastRestartRequestTime sets LastRestartRequestTime field to given value.

### HasLastRestartRequestTime

`func (o *LifecycleConfig) HasLastRestartRequestTime() bool`

HasLastRestartRequestTime returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


