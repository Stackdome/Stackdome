# RegistryCredential

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**Host** | **string** | Registry host; normalized on create (docker.io aliases become index.docker.io) | 
**Purpose** | Pointer to [**RegistryCredentialPurpose**](RegistryCredentialPurpose.md) |  | [optional] [default to BOTH]
**Username** | **string** |  | 
**Password** | Pointer to **string** | Required on create; omit on update to keep the stored password | [optional] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewRegistryCredential

`func NewRegistryCredential(host string, username string, ) *RegistryCredential`

NewRegistryCredential instantiates a new RegistryCredential object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegistryCredentialWithDefaults

`func NewRegistryCredentialWithDefaults() *RegistryCredential`

NewRegistryCredentialWithDefaults instantiates a new RegistryCredential object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *RegistryCredential) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RegistryCredential) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RegistryCredential) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *RegistryCredential) HasId() bool`

HasId returns a boolean if a field has been set.

### GetHost

`func (o *RegistryCredential) GetHost() string`

GetHost returns the Host field if non-nil, zero value otherwise.

### GetHostOk

`func (o *RegistryCredential) GetHostOk() (*string, bool)`

GetHostOk returns a tuple with the Host field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHost

`func (o *RegistryCredential) SetHost(v string)`

SetHost sets Host field to given value.


### GetPurpose

`func (o *RegistryCredential) GetPurpose() RegistryCredentialPurpose`

GetPurpose returns the Purpose field if non-nil, zero value otherwise.

### GetPurposeOk

`func (o *RegistryCredential) GetPurposeOk() (*RegistryCredentialPurpose, bool)`

GetPurposeOk returns a tuple with the Purpose field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPurpose

`func (o *RegistryCredential) SetPurpose(v RegistryCredentialPurpose)`

SetPurpose sets Purpose field to given value.

### HasPurpose

`func (o *RegistryCredential) HasPurpose() bool`

HasPurpose returns a boolean if a field has been set.

### GetUsername

`func (o *RegistryCredential) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *RegistryCredential) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *RegistryCredential) SetUsername(v string)`

SetUsername sets Username field to given value.


### GetPassword

`func (o *RegistryCredential) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *RegistryCredential) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *RegistryCredential) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *RegistryCredential) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetOrganisationId

`func (o *RegistryCredential) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *RegistryCredential) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *RegistryCredential) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *RegistryCredential) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetCreatedAt

`func (o *RegistryCredential) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *RegistryCredential) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *RegistryCredential) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *RegistryCredential) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *RegistryCredential) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *RegistryCredential) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *RegistryCredential) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *RegistryCredential) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


