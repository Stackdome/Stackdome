# PostgresAddon

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**UserId** | Pointer to **string** |  | [optional] [readonly] 
**ClusterId** | Pointer to **string** |  | [optional] [readonly] 
**Name** | **string** | Unique name for this PostgreSQL cluster | 
**Namespace** | Pointer to **string** |  | [optional] [readonly] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 
**Revision** | Pointer to **string** |  | [optional] [readonly] 
**Spec** | [**PostgresAddonSpec**](PostgresAddonSpec.md) |  | 
**Status** | Pointer to [**PostgresAddonStatus**](PostgresAddonStatus.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewPostgresAddon

`func NewPostgresAddon(name string, spec PostgresAddonSpec, ) *PostgresAddon`

NewPostgresAddon instantiates a new PostgresAddon object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresAddonWithDefaults

`func NewPostgresAddonWithDefaults() *PostgresAddon`

NewPostgresAddonWithDefaults instantiates a new PostgresAddon object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *PostgresAddon) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PostgresAddon) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PostgresAddon) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *PostgresAddon) HasId() bool`

HasId returns a boolean if a field has been set.

### GetOrganisationId

`func (o *PostgresAddon) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *PostgresAddon) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *PostgresAddon) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *PostgresAddon) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetUserId

`func (o *PostgresAddon) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *PostgresAddon) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *PostgresAddon) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *PostgresAddon) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetClusterId

`func (o *PostgresAddon) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *PostgresAddon) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *PostgresAddon) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *PostgresAddon) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetName

`func (o *PostgresAddon) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PostgresAddon) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PostgresAddon) SetName(v string)`

SetName sets Name field to given value.


### GetNamespace

`func (o *PostgresAddon) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *PostgresAddon) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *PostgresAddon) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *PostgresAddon) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetLabels

`func (o *PostgresAddon) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *PostgresAddon) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *PostgresAddon) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *PostgresAddon) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *PostgresAddon) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *PostgresAddon) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *PostgresAddon) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *PostgresAddon) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetRevision

`func (o *PostgresAddon) GetRevision() string`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *PostgresAddon) GetRevisionOk() (*string, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *PostgresAddon) SetRevision(v string)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *PostgresAddon) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### GetSpec

`func (o *PostgresAddon) GetSpec() PostgresAddonSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *PostgresAddon) GetSpecOk() (*PostgresAddonSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *PostgresAddon) SetSpec(v PostgresAddonSpec)`

SetSpec sets Spec field to given value.


### GetStatus

`func (o *PostgresAddon) GetStatus() PostgresAddonStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *PostgresAddon) GetStatusOk() (*PostgresAddonStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *PostgresAddon) SetStatus(v PostgresAddonStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *PostgresAddon) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreatedAt

`func (o *PostgresAddon) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *PostgresAddon) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *PostgresAddon) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *PostgresAddon) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *PostgresAddon) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *PostgresAddon) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *PostgresAddon) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *PostgresAddon) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


