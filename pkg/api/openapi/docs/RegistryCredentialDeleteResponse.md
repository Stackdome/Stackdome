# RegistryCredentialDeleteResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AffectedStacks** | Pointer to [**[]AffectedStackRef**](AffectedStackRef.md) | Stacks that were implicitly resolving against the deleted credential | [optional] 

## Methods

### NewRegistryCredentialDeleteResponse

`func NewRegistryCredentialDeleteResponse() *RegistryCredentialDeleteResponse`

NewRegistryCredentialDeleteResponse instantiates a new RegistryCredentialDeleteResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryCredentialDeleteResponseWithDefaults

`func NewRegistryCredentialDeleteResponseWithDefaults() *RegistryCredentialDeleteResponse`

NewRegistryCredentialDeleteResponseWithDefaults instantiates a new RegistryCredentialDeleteResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAffectedStacks

`func (o *RegistryCredentialDeleteResponse) GetAffectedStacks() []AffectedStackRef`

GetAffectedStacks returns the AffectedStacks field if non-nil, zero value otherwise.

### GetAffectedStacksOk

`func (o *RegistryCredentialDeleteResponse) GetAffectedStacksOk() (*[]AffectedStackRef, bool)`

GetAffectedStacksOk returns a tuple with the AffectedStacks field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAffectedStacks

`func (o *RegistryCredentialDeleteResponse) SetAffectedStacks(v []AffectedStackRef)`

SetAffectedStacks sets AffectedStacks field to given value.

### HasAffectedStacks

`func (o *RegistryCredentialDeleteResponse) HasAffectedStacks() bool`

HasAffectedStacks returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


