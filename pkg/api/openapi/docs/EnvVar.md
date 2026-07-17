# EnvVar

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**Value** | Pointer to **string** | Literal environment variable value. | [optional] 
**SelfOutput** | Pointer to **string** | Read this environment variable from one of the resource&#39;s own declared outputs, for example public_url. | [optional] 

## Methods

### NewEnvVar

`func NewEnvVar(name string, ) *EnvVar`

NewEnvVar instantiates a new EnvVar object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvVarWithDefaults

`func NewEnvVarWithDefaults() *EnvVar`

NewEnvVarWithDefaults instantiates a new EnvVar object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *EnvVar) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *EnvVar) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *EnvVar) SetName(v string)`

SetName sets Name field to given value.


### GetValue

`func (o *EnvVar) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *EnvVar) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *EnvVar) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *EnvVar) HasValue() bool`

HasValue returns a boolean if a field has been set.

### GetSelfOutput

`func (o *EnvVar) GetSelfOutput() string`

GetSelfOutput returns the SelfOutput field if non-nil, zero value otherwise.

### GetSelfOutputOk

`func (o *EnvVar) GetSelfOutputOk() (*string, bool)`

GetSelfOutputOk returns a tuple with the SelfOutput field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelfOutput

`func (o *EnvVar) SetSelfOutput(v string)`

SetSelfOutput sets SelfOutput field to given value.

### HasSelfOutput

`func (o *EnvVar) HasSelfOutput() bool`

HasSelfOutput returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


