# RemoteSyncServer

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**Name** | **string** |  | 
**Namespace** | Pointer to **string** |  | [optional] [readonly] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 
**Spec** | [**RemoteSyncServerSpec**](RemoteSyncServerSpec.md) |  | 
**Status** | Pointer to [**RemoteSyncServerStatus**](RemoteSyncServerStatus.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewRemoteSyncServer

`func NewRemoteSyncServer(name string, spec RemoteSyncServerSpec, ) *RemoteSyncServer`

NewRemoteSyncServer instantiates a new RemoteSyncServer object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRemoteSyncServerWithDefaults

`func NewRemoteSyncServerWithDefaults() *RemoteSyncServer`

NewRemoteSyncServerWithDefaults instantiates a new RemoteSyncServer object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *RemoteSyncServer) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *RemoteSyncServer) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *RemoteSyncServer) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *RemoteSyncServer) HasId() bool`

HasId returns a boolean if a field has been set.

### GetOrganisationId

`func (o *RemoteSyncServer) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *RemoteSyncServer) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *RemoteSyncServer) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *RemoteSyncServer) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetName

`func (o *RemoteSyncServer) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *RemoteSyncServer) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *RemoteSyncServer) SetName(v string)`

SetName sets Name field to given value.


### GetNamespace

`func (o *RemoteSyncServer) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *RemoteSyncServer) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *RemoteSyncServer) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *RemoteSyncServer) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetLabels

`func (o *RemoteSyncServer) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *RemoteSyncServer) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *RemoteSyncServer) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *RemoteSyncServer) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *RemoteSyncServer) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *RemoteSyncServer) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *RemoteSyncServer) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *RemoteSyncServer) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetSpec

`func (o *RemoteSyncServer) GetSpec() RemoteSyncServerSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *RemoteSyncServer) GetSpecOk() (*RemoteSyncServerSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *RemoteSyncServer) SetSpec(v RemoteSyncServerSpec)`

SetSpec sets Spec field to given value.


### GetStatus

`func (o *RemoteSyncServer) GetStatus() RemoteSyncServerStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *RemoteSyncServer) GetStatusOk() (*RemoteSyncServerStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *RemoteSyncServer) SetStatus(v RemoteSyncServerStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *RemoteSyncServer) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreatedAt

`func (o *RemoteSyncServer) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *RemoteSyncServer) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *RemoteSyncServer) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *RemoteSyncServer) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *RemoteSyncServer) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *RemoteSyncServer) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *RemoteSyncServer) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *RemoteSyncServer) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


