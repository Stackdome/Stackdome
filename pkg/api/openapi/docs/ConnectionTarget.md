# ConnectionTarget

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | env writes an environment variable and file writes a mounted file path. | 
**Name** | Pointer to **string** | Environment variable name when type is env. | [optional] 
**Path** | Pointer to **string** | Absolute file path when type is file. | [optional] 

## Methods

### NewConnectionTarget

`func NewConnectionTarget(type_ string, ) *ConnectionTarget`

NewConnectionTarget instantiates a new ConnectionTarget object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectionTargetWithDefaults

`func NewConnectionTargetWithDefaults() *ConnectionTarget`

NewConnectionTargetWithDefaults instantiates a new ConnectionTarget object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *ConnectionTarget) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ConnectionTarget) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ConnectionTarget) SetType(v string)`

SetType sets Type field to given value.


### GetName

`func (o *ConnectionTarget) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ConnectionTarget) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ConnectionTarget) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ConnectionTarget) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPath

`func (o *ConnectionTarget) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *ConnectionTarget) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *ConnectionTarget) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *ConnectionTarget) HasPath() bool`

HasPath returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


