# Port

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Number** | **int32** |  | 
**Protocol** | Pointer to **string** |  | [optional] 
**ExposedToPublic** | **bool** |  | 
**SubdomainPrefix** | Pointer to **string** |  | [optional] 

## Methods

### NewPort

`func NewPort(number int32, exposedToPublic bool, ) *Port`

NewPort instantiates a new Port object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPortWithDefaults

`func NewPortWithDefaults() *Port`

NewPortWithDefaults instantiates a new Port object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNumber

`func (o *Port) GetNumber() int32`

GetNumber returns the Number field if non-nil, zero value otherwise.

### GetNumberOk

`func (o *Port) GetNumberOk() (*int32, bool)`

GetNumberOk returns a tuple with the Number field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumber

`func (o *Port) SetNumber(v int32)`

SetNumber sets Number field to given value.


### GetProtocol

`func (o *Port) GetProtocol() string`

GetProtocol returns the Protocol field if non-nil, zero value otherwise.

### GetProtocolOk

`func (o *Port) GetProtocolOk() (*string, bool)`

GetProtocolOk returns a tuple with the Protocol field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProtocol

`func (o *Port) SetProtocol(v string)`

SetProtocol sets Protocol field to given value.

### HasProtocol

`func (o *Port) HasProtocol() bool`

HasProtocol returns a boolean if a field has been set.

### GetExposedToPublic

`func (o *Port) GetExposedToPublic() bool`

GetExposedToPublic returns the ExposedToPublic field if non-nil, zero value otherwise.

### GetExposedToPublicOk

`func (o *Port) GetExposedToPublicOk() (*bool, bool)`

GetExposedToPublicOk returns a tuple with the ExposedToPublic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedToPublic

`func (o *Port) SetExposedToPublic(v bool)`

SetExposedToPublic sets ExposedToPublic field to given value.


### GetSubdomainPrefix

`func (o *Port) GetSubdomainPrefix() string`

GetSubdomainPrefix returns the SubdomainPrefix field if non-nil, zero value otherwise.

### GetSubdomainPrefixOk

`func (o *Port) GetSubdomainPrefixOk() (*string, bool)`

GetSubdomainPrefixOk returns a tuple with the SubdomainPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubdomainPrefix

`func (o *Port) SetSubdomainPrefix(v string)`

SetSubdomainPrefix sets SubdomainPrefix field to given value.

### HasSubdomainPrefix

`func (o *Port) HasSubdomainPrefix() bool`

HasSubdomainPrefix returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


