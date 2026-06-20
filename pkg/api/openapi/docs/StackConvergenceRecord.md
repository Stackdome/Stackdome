# StackConvergenceRecord

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Revision** | Pointer to **string** |  | [optional] 
**ReleaseId** | Pointer to **string** |  | [optional] 
**At** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewStackConvergenceRecord

`func NewStackConvergenceRecord() *StackConvergenceRecord`

NewStackConvergenceRecord instantiates a new StackConvergenceRecord object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackConvergenceRecordWithDefaults

`func NewStackConvergenceRecordWithDefaults() *StackConvergenceRecord`

NewStackConvergenceRecordWithDefaults instantiates a new StackConvergenceRecord object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRevision

`func (o *StackConvergenceRecord) GetRevision() string`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *StackConvergenceRecord) GetRevisionOk() (*string, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *StackConvergenceRecord) SetRevision(v string)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *StackConvergenceRecord) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### GetReleaseId

`func (o *StackConvergenceRecord) GetReleaseId() string`

GetReleaseId returns the ReleaseId field if non-nil, zero value otherwise.

### GetReleaseIdOk

`func (o *StackConvergenceRecord) GetReleaseIdOk() (*string, bool)`

GetReleaseIdOk returns a tuple with the ReleaseId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReleaseId

`func (o *StackConvergenceRecord) SetReleaseId(v string)`

SetReleaseId sets ReleaseId field to given value.

### HasReleaseId

`func (o *StackConvergenceRecord) HasReleaseId() bool`

HasReleaseId returns a boolean if a field has been set.

### GetAt

`func (o *StackConvergenceRecord) GetAt() time.Time`

GetAt returns the At field if non-nil, zero value otherwise.

### GetAtOk

`func (o *StackConvergenceRecord) GetAtOk() (*time.Time, bool)`

GetAtOk returns a tuple with the At field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAt

`func (o *StackConvergenceRecord) SetAt(v time.Time)`

SetAt sets At field to given value.

### HasAt

`func (o *StackConvergenceRecord) HasAt() bool`

HasAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


