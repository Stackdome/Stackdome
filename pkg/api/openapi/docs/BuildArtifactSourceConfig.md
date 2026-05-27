# BuildArtifactSourceConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourcePath** | **string** | Path within the build output to copy from. | 
**DestinationPath** | Pointer to **string** | Path within the volume to copy to. | [optional] 

## Methods

### NewBuildArtifactSourceConfig

`func NewBuildArtifactSourceConfig(sourcePath string, ) *BuildArtifactSourceConfig`

NewBuildArtifactSourceConfig instantiates a new BuildArtifactSourceConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBuildArtifactSourceConfigWithDefaults

`func NewBuildArtifactSourceConfigWithDefaults() *BuildArtifactSourceConfig`

NewBuildArtifactSourceConfigWithDefaults instantiates a new BuildArtifactSourceConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourcePath

`func (o *BuildArtifactSourceConfig) GetSourcePath() string`

GetSourcePath returns the SourcePath field if non-nil, zero value otherwise.

### GetSourcePathOk

`func (o *BuildArtifactSourceConfig) GetSourcePathOk() (*string, bool)`

GetSourcePathOk returns a tuple with the SourcePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourcePath

`func (o *BuildArtifactSourceConfig) SetSourcePath(v string)`

SetSourcePath sets SourcePath field to given value.


### GetDestinationPath

`func (o *BuildArtifactSourceConfig) GetDestinationPath() string`

GetDestinationPath returns the DestinationPath field if non-nil, zero value otherwise.

### GetDestinationPathOk

`func (o *BuildArtifactSourceConfig) GetDestinationPathOk() (*string, bool)`

GetDestinationPathOk returns a tuple with the DestinationPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDestinationPath

`func (o *BuildArtifactSourceConfig) SetDestinationPath(v string)`

SetDestinationPath sets DestinationPath field to given value.

### HasDestinationPath

`func (o *BuildArtifactSourceConfig) HasDestinationPath() bool`

HasDestinationPath returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


