# Organisation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Domains** | Pointer to [**[]DomainName**](DomainName.md) |  | [optional] 
**IsPlatform** | Pointer to **bool** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewOrganisation

`func NewOrganisation() *Organisation`

NewOrganisation instantiates a new Organisation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOrganisationWithDefaults

`func NewOrganisationWithDefaults() *Organisation`

NewOrganisationWithDefaults instantiates a new Organisation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Organisation) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Organisation) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Organisation) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *Organisation) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *Organisation) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Organisation) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Organisation) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *Organisation) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDomains

`func (o *Organisation) GetDomains() []DomainName`

GetDomains returns the Domains field if non-nil, zero value otherwise.

### GetDomainsOk

`func (o *Organisation) GetDomainsOk() (*[]DomainName, bool)`

GetDomainsOk returns a tuple with the Domains field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomains

`func (o *Organisation) SetDomains(v []DomainName)`

SetDomains sets Domains field to given value.

### HasDomains

`func (o *Organisation) HasDomains() bool`

HasDomains returns a boolean if a field has been set.

### GetIsPlatform

`func (o *Organisation) GetIsPlatform() bool`

GetIsPlatform returns the IsPlatform field if non-nil, zero value otherwise.

### GetIsPlatformOk

`func (o *Organisation) GetIsPlatformOk() (*bool, bool)`

GetIsPlatformOk returns a tuple with the IsPlatform field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsPlatform

`func (o *Organisation) SetIsPlatform(v bool)`

SetIsPlatform sets IsPlatform field to given value.

### HasIsPlatform

`func (o *Organisation) HasIsPlatform() bool`

HasIsPlatform returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Organisation) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Organisation) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Organisation) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Organisation) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *Organisation) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Organisation) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Organisation) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *Organisation) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


