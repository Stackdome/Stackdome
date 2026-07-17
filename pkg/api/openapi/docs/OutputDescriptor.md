# OutputDescriptor

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Stable output accessor name, for example &#x60;host&#x60; or &#x60;public_url&#x60;. | 
**Type** | **string** | Scalar value type exposed by this output. | 
**Sensitive** | **bool** | True when the output value is sensitive and should never be returned in normal metadata APIs. | 

## Methods

### NewOutputDescriptor

`func NewOutputDescriptor(name string, type_ string, sensitive bool, ) *OutputDescriptor`

NewOutputDescriptor instantiates a new OutputDescriptor object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOutputDescriptorWithDefaults

`func NewOutputDescriptorWithDefaults() *OutputDescriptor`

NewOutputDescriptorWithDefaults instantiates a new OutputDescriptor object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *OutputDescriptor) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *OutputDescriptor) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *OutputDescriptor) SetName(v string)`

SetName sets Name field to given value.


### GetType

`func (o *OutputDescriptor) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *OutputDescriptor) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *OutputDescriptor) SetType(v string)`

SetType sets Type field to given value.


### GetSensitive

`func (o *OutputDescriptor) GetSensitive() bool`

GetSensitive returns the Sensitive field if non-nil, zero value otherwise.

### GetSensitiveOk

`func (o *OutputDescriptor) GetSensitiveOk() (*bool, bool)`

GetSensitiveOk returns a tuple with the Sensitive field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSensitive

`func (o *OutputDescriptor) SetSensitive(v bool)`

SetSensitive sets Sensitive field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


