# LocalSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Path** | **string** |  | 
**Sync** | **bool** |  | 

## Methods

### NewLocalSource

`func NewLocalSource(path string, sync bool, ) *LocalSource`

NewLocalSource instantiates a new LocalSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLocalSourceWithDefaults

`func NewLocalSourceWithDefaults() *LocalSource`

NewLocalSourceWithDefaults instantiates a new LocalSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPath

`func (o *LocalSource) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *LocalSource) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *LocalSource) SetPath(v string)`

SetPath sets Path field to given value.


### GetSync

`func (o *LocalSource) GetSync() bool`

GetSync returns the Sync field if non-nil, zero value otherwise.

### GetSyncOk

`func (o *LocalSource) GetSyncOk() (*bool, bool)`

GetSyncOk returns a tuple with the Sync field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSync

`func (o *LocalSource) SetSync(v bool)`

SetSync sets Sync field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


