# WorkspaceStorage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**Name** | **string** |  | 
**Namespace** | **string** |  | 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 
**SshConfig** | [**SSHConfig**](SSHConfig.md) |  | 
**Volumes** | [**[]Volume**](Volume.md) |  | 
**Status** | Pointer to [**WorkspaceStorageStatus**](WorkspaceStorageStatus.md) |  | [optional] 
**State** | Pointer to [**WorkspaceStorageState**](WorkspaceStorageState.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewWorkspaceStorage

`func NewWorkspaceStorage(name string, namespace string, sshConfig SSHConfig, volumes []Volume, ) *WorkspaceStorage`

NewWorkspaceStorage instantiates a new WorkspaceStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceStorageWithDefaults

`func NewWorkspaceStorageWithDefaults() *WorkspaceStorage`

NewWorkspaceStorageWithDefaults instantiates a new WorkspaceStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkspaceStorage) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkspaceStorage) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkspaceStorage) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *WorkspaceStorage) HasId() bool`

HasId returns a boolean if a field has been set.

### GetOrganisationId

`func (o *WorkspaceStorage) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *WorkspaceStorage) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *WorkspaceStorage) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *WorkspaceStorage) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetName

`func (o *WorkspaceStorage) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkspaceStorage) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkspaceStorage) SetName(v string)`

SetName sets Name field to given value.


### GetNamespace

`func (o *WorkspaceStorage) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *WorkspaceStorage) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *WorkspaceStorage) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetLabels

`func (o *WorkspaceStorage) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *WorkspaceStorage) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *WorkspaceStorage) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *WorkspaceStorage) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *WorkspaceStorage) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *WorkspaceStorage) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *WorkspaceStorage) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *WorkspaceStorage) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetSshConfig

`func (o *WorkspaceStorage) GetSshConfig() SSHConfig`

GetSshConfig returns the SshConfig field if non-nil, zero value otherwise.

### GetSshConfigOk

`func (o *WorkspaceStorage) GetSshConfigOk() (*SSHConfig, bool)`

GetSshConfigOk returns a tuple with the SshConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshConfig

`func (o *WorkspaceStorage) SetSshConfig(v SSHConfig)`

SetSshConfig sets SshConfig field to given value.


### GetVolumes

`func (o *WorkspaceStorage) GetVolumes() []Volume`

GetVolumes returns the Volumes field if non-nil, zero value otherwise.

### GetVolumesOk

`func (o *WorkspaceStorage) GetVolumesOk() (*[]Volume, bool)`

GetVolumesOk returns a tuple with the Volumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumes

`func (o *WorkspaceStorage) SetVolumes(v []Volume)`

SetVolumes sets Volumes field to given value.


### GetStatus

`func (o *WorkspaceStorage) GetStatus() WorkspaceStorageStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *WorkspaceStorage) GetStatusOk() (*WorkspaceStorageStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *WorkspaceStorage) SetStatus(v WorkspaceStorageStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *WorkspaceStorage) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetState

`func (o *WorkspaceStorage) GetState() WorkspaceStorageState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *WorkspaceStorage) GetStateOk() (*WorkspaceStorageState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *WorkspaceStorage) SetState(v WorkspaceStorageState)`

SetState sets State field to given value.

### HasState

`func (o *WorkspaceStorage) HasState() bool`

HasState returns a boolean if a field has been set.

### GetCreatedAt

`func (o *WorkspaceStorage) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkspaceStorage) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkspaceStorage) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkspaceStorage) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkspaceStorage) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkspaceStorage) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkspaceStorage) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkspaceStorage) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


