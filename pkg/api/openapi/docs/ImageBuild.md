# ImageBuild

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Namespace** | Pointer to **string** |  | [optional] 
**StackId** | Pointer to **string** |  | [optional] 
**StackResourceId** | **string** |  | 
**StackResourceName** | **string** |  | 
**SourceRevision** | [**BuildSourceRevision**](BuildSourceRevision.md) |  | 
**BuildContext** | [**BuildSourceContext**](BuildSourceContext.md) |  | 
**ImageRepo** | **string** |  | 
**Status** | Pointer to [**ImageBuildStatus**](ImageBuildStatus.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewImageBuild

`func NewImageBuild(stackResourceId string, stackResourceName string, sourceRevision BuildSourceRevision, buildContext BuildSourceContext, imageRepo string, ) *ImageBuild`

NewImageBuild instantiates a new ImageBuild object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImageBuildWithDefaults

`func NewImageBuildWithDefaults() *ImageBuild`

NewImageBuildWithDefaults instantiates a new ImageBuild object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ImageBuild) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ImageBuild) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ImageBuild) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ImageBuild) HasId() bool`

HasId returns a boolean if a field has been set.

### GetNamespace

`func (o *ImageBuild) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ImageBuild) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ImageBuild) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ImageBuild) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetStackId

`func (o *ImageBuild) GetStackId() string`

GetStackId returns the StackId field if non-nil, zero value otherwise.

### GetStackIdOk

`func (o *ImageBuild) GetStackIdOk() (*string, bool)`

GetStackIdOk returns a tuple with the StackId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackId

`func (o *ImageBuild) SetStackId(v string)`

SetStackId sets StackId field to given value.

### HasStackId

`func (o *ImageBuild) HasStackId() bool`

HasStackId returns a boolean if a field has been set.

### GetStackResourceId

`func (o *ImageBuild) GetStackResourceId() string`

GetStackResourceId returns the StackResourceId field if non-nil, zero value otherwise.

### GetStackResourceIdOk

`func (o *ImageBuild) GetStackResourceIdOk() (*string, bool)`

GetStackResourceIdOk returns a tuple with the StackResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackResourceId

`func (o *ImageBuild) SetStackResourceId(v string)`

SetStackResourceId sets StackResourceId field to given value.


### GetStackResourceName

`func (o *ImageBuild) GetStackResourceName() string`

GetStackResourceName returns the StackResourceName field if non-nil, zero value otherwise.

### GetStackResourceNameOk

`func (o *ImageBuild) GetStackResourceNameOk() (*string, bool)`

GetStackResourceNameOk returns a tuple with the StackResourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackResourceName

`func (o *ImageBuild) SetStackResourceName(v string)`

SetStackResourceName sets StackResourceName field to given value.


### GetSourceRevision

`func (o *ImageBuild) GetSourceRevision() BuildSourceRevision`

GetSourceRevision returns the SourceRevision field if non-nil, zero value otherwise.

### GetSourceRevisionOk

`func (o *ImageBuild) GetSourceRevisionOk() (*BuildSourceRevision, bool)`

GetSourceRevisionOk returns a tuple with the SourceRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceRevision

`func (o *ImageBuild) SetSourceRevision(v BuildSourceRevision)`

SetSourceRevision sets SourceRevision field to given value.


### GetBuildContext

`func (o *ImageBuild) GetBuildContext() BuildSourceContext`

GetBuildContext returns the BuildContext field if non-nil, zero value otherwise.

### GetBuildContextOk

`func (o *ImageBuild) GetBuildContextOk() (*BuildSourceContext, bool)`

GetBuildContextOk returns a tuple with the BuildContext field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildContext

`func (o *ImageBuild) SetBuildContext(v BuildSourceContext)`

SetBuildContext sets BuildContext field to given value.


### GetImageRepo

`func (o *ImageBuild) GetImageRepo() string`

GetImageRepo returns the ImageRepo field if non-nil, zero value otherwise.

### GetImageRepoOk

`func (o *ImageBuild) GetImageRepoOk() (*string, bool)`

GetImageRepoOk returns a tuple with the ImageRepo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageRepo

`func (o *ImageBuild) SetImageRepo(v string)`

SetImageRepo sets ImageRepo field to given value.


### GetStatus

`func (o *ImageBuild) GetStatus() ImageBuildStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ImageBuild) GetStatusOk() (*ImageBuildStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ImageBuild) SetStatus(v ImageBuildStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ImageBuild) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ImageBuild) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ImageBuild) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ImageBuild) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ImageBuild) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *ImageBuild) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *ImageBuild) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *ImageBuild) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *ImageBuild) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


