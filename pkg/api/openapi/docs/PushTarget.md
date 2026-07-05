# PushTarget

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Repository** | **string** | Push repository as host/path, without a tag | 
**RegistryCredentialsId** | Pointer to **string** | Org-level registry credential override for push auth | [optional] 

## Methods

### NewPushTarget

`func NewPushTarget(repository string, ) *PushTarget`

NewPushTarget instantiates a new PushTarget object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPushTargetWithDefaults

`func NewPushTargetWithDefaults() *PushTarget`

NewPushTargetWithDefaults instantiates a new PushTarget object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepository

`func (o *PushTarget) GetRepository() string`

GetRepository returns the Repository field if non-nil, zero value otherwise.

### GetRepositoryOk

`func (o *PushTarget) GetRepositoryOk() (*string, bool)`

GetRepositoryOk returns a tuple with the Repository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepository

`func (o *PushTarget) SetRepository(v string)`

SetRepository sets Repository field to given value.


### GetRegistryCredentialsId

`func (o *PushTarget) GetRegistryCredentialsId() string`

GetRegistryCredentialsId returns the RegistryCredentialsId field if non-nil, zero value otherwise.

### GetRegistryCredentialsIdOk

`func (o *PushTarget) GetRegistryCredentialsIdOk() (*string, bool)`

GetRegistryCredentialsIdOk returns a tuple with the RegistryCredentialsId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegistryCredentialsId

`func (o *PushTarget) SetRegistryCredentialsId(v string)`

SetRegistryCredentialsId sets RegistryCredentialsId field to given value.

### HasRegistryCredentialsId

`func (o *PushTarget) HasRegistryCredentialsId() bool`

HasRegistryCredentialsId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


