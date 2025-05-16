# StackResourceBuildSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceContext** | [**BuildSourceContext**](BuildSourceContext.md) |  | 
**ContextPathWithinSource** | **string** |  | 
**DockerfilePath** | **string** |  | 
**SourceRevision** | [**BuildSourceRevision**](BuildSourceRevision.md) |  | 
**ImageRepository** | Pointer to [**ImageRepository**](ImageRepository.md) |  | [optional] 

## Methods

### NewStackResourceBuildSpec

`func NewStackResourceBuildSpec(sourceContext BuildSourceContext, contextPathWithinSource string, dockerfilePath string, sourceRevision BuildSourceRevision, ) *StackResourceBuildSpec`

NewStackResourceBuildSpec instantiates a new StackResourceBuildSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackResourceBuildSpecWithDefaults

`func NewStackResourceBuildSpecWithDefaults() *StackResourceBuildSpec`

NewStackResourceBuildSpecWithDefaults instantiates a new StackResourceBuildSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceContext

`func (o *StackResourceBuildSpec) GetSourceContext() BuildSourceContext`

GetSourceContext returns the SourceContext field if non-nil, zero value otherwise.

### GetSourceContextOk

`func (o *StackResourceBuildSpec) GetSourceContextOk() (*BuildSourceContext, bool)`

GetSourceContextOk returns a tuple with the SourceContext field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceContext

`func (o *StackResourceBuildSpec) SetSourceContext(v BuildSourceContext)`

SetSourceContext sets SourceContext field to given value.


### GetContextPathWithinSource

`func (o *StackResourceBuildSpec) GetContextPathWithinSource() string`

GetContextPathWithinSource returns the ContextPathWithinSource field if non-nil, zero value otherwise.

### GetContextPathWithinSourceOk

`func (o *StackResourceBuildSpec) GetContextPathWithinSourceOk() (*string, bool)`

GetContextPathWithinSourceOk returns a tuple with the ContextPathWithinSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContextPathWithinSource

`func (o *StackResourceBuildSpec) SetContextPathWithinSource(v string)`

SetContextPathWithinSource sets ContextPathWithinSource field to given value.


### GetDockerfilePath

`func (o *StackResourceBuildSpec) GetDockerfilePath() string`

GetDockerfilePath returns the DockerfilePath field if non-nil, zero value otherwise.

### GetDockerfilePathOk

`func (o *StackResourceBuildSpec) GetDockerfilePathOk() (*string, bool)`

GetDockerfilePathOk returns a tuple with the DockerfilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDockerfilePath

`func (o *StackResourceBuildSpec) SetDockerfilePath(v string)`

SetDockerfilePath sets DockerfilePath field to given value.


### GetSourceRevision

`func (o *StackResourceBuildSpec) GetSourceRevision() BuildSourceRevision`

GetSourceRevision returns the SourceRevision field if non-nil, zero value otherwise.

### GetSourceRevisionOk

`func (o *StackResourceBuildSpec) GetSourceRevisionOk() (*BuildSourceRevision, bool)`

GetSourceRevisionOk returns a tuple with the SourceRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceRevision

`func (o *StackResourceBuildSpec) SetSourceRevision(v BuildSourceRevision)`

SetSourceRevision sets SourceRevision field to given value.


### GetImageRepository

`func (o *StackResourceBuildSpec) GetImageRepository() ImageRepository`

GetImageRepository returns the ImageRepository field if non-nil, zero value otherwise.

### GetImageRepositoryOk

`func (o *StackResourceBuildSpec) GetImageRepositoryOk() (*ImageRepository, bool)`

GetImageRepositoryOk returns a tuple with the ImageRepository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageRepository

`func (o *StackResourceBuildSpec) SetImageRepository(v ImageRepository)`

SetImageRepository sets ImageRepository field to given value.

### HasImageRepository

`func (o *StackResourceBuildSpec) HasImageRepository() bool`

HasImageRepository returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


