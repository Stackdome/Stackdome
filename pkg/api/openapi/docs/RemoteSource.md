# RemoteSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | **string** |  | 
**CurrentDirectoryHash** | **string** |  | 

## Methods

### NewRemoteSource

`func NewRemoteSource(path string, currentDirectoryHash string, ) *RemoteSource`

NewRemoteSource instantiates a new RemoteSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRemoteSourceWithDefaults

`func NewRemoteSourceWithDefaults() *RemoteSource`

NewRemoteSourceWithDefaults instantiates a new RemoteSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *RemoteSource) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *RemoteSource) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *RemoteSource) SetPath(v string)`

SetPath sets Path field to given value.


### GetCurrentDirectoryHash

`func (o *RemoteSource) GetCurrentDirectoryHash() string`

GetCurrentDirectoryHash returns the CurrentDirectoryHash field if non-nil, zero value otherwise.

### GetCurrentDirectoryHashOk

`func (o *RemoteSource) GetCurrentDirectoryHashOk() (*string, bool)`

GetCurrentDirectoryHashOk returns a tuple with the CurrentDirectoryHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentDirectoryHash

`func (o *RemoteSource) SetCurrentDirectoryHash(v string)`

SetCurrentDirectoryHash sets CurrentDirectoryHash field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


