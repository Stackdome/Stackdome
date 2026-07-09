# RegistryCredentialVerifyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Repository** | **string** | Repository to verify against, e.g. ghcr.io/acme/app | 
**Purpose** | Pointer to [**RegistryCredentialPurpose**](RegistryCredentialPurpose.md) |  | [optional] [default to BOTH]

## Methods

### NewRegistryCredentialVerifyRequest

`func NewRegistryCredentialVerifyRequest(repository string, ) *RegistryCredentialVerifyRequest`

NewRegistryCredentialVerifyRequest instantiates a new RegistryCredentialVerifyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryCredentialVerifyRequestWithDefaults

`func NewRegistryCredentialVerifyRequestWithDefaults() *RegistryCredentialVerifyRequest`

NewRegistryCredentialVerifyRequestWithDefaults instantiates a new RegistryCredentialVerifyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepository

`func (o *RegistryCredentialVerifyRequest) GetRepository() string`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *RegistryCredentialVerifyRequest) GetRepositoryOk() (*string, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *RegistryCredentialVerifyRequest) SetRepository(v string)`

SetRepository sets Repository field to given value.


### GetPurpose

`func (o *RegistryCredentialVerifyRequest) GetPurpose() RegistryCredentialPurpose`

GetPurpose returns the Purpose field if non-nil, zero value otherwise.

### GetPurposeOk

`func (o *RegistryCredentialVerifyRequest) GetPurposeOk() (*RegistryCredentialPurpose, bool)`

GetPurposeOk returns a tuple with the Purpose field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPurpose

`func (o *RegistryCredentialVerifyRequest) SetPurpose(v RegistryCredentialPurpose)`

SetPurpose sets Purpose field to given value.

### HasPurpose

`func (o *RegistryCredentialVerifyRequest) HasPurpose() bool`

HasPurpose returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


