# BuildConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ContextPath** | **string** |  | 
**DockerfilePath** | **string** |  | 
**SourceHash** | **string** |  | 

## Methods

### NewBuildConfig

`func NewBuildConfig(contextPath string, dockerfilePath string, sourceHash string, ) *BuildConfig`

NewBuildConfig instantiates a new BuildConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBuildConfigWithDefaults

`func NewBuildConfigWithDefaults() *BuildConfig`

NewBuildConfigWithDefaults instantiates a new BuildConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContextPath

`func (o *BuildConfig) GetContextPath() string`

GetContextPath returns the ContextPath field if non-nil, zero value otherwise.

### GetContextPathOk

`func (o *BuildConfig) GetContextPathOk() (*string, bool)`

GetContextPathOk returns a tuple with the ContextPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContextPath

`func (o *BuildConfig) SetContextPath(v string)`

SetContextPath sets ContextPath field to given value.


### GetDockerfilePath

`func (o *BuildConfig) GetDockerfilePath() string`

GetDockerfilePath returns the DockerfilePath field if non-nil, zero value otherwise.

### GetDockerfilePathOk

`func (o *BuildConfig) GetDockerfilePathOk() (*string, bool)`

GetDockerfilePathOk returns a tuple with the DockerfilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDockerfilePath

`func (o *BuildConfig) SetDockerfilePath(v string)`

SetDockerfilePath sets DockerfilePath field to given value.


### GetSourceHash

`func (o *BuildConfig) GetSourceHash() string`

GetSourceHash returns the SourceHash field if non-nil, zero value otherwise.

### GetSourceHashOk

`func (o *BuildConfig) GetSourceHashOk() (*string, bool)`

GetSourceHashOk returns a tuple with the SourceHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceHash

`func (o *BuildConfig) SetSourceHash(v string)`

SetSourceHash sets SourceHash field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


