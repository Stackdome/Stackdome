# ReleaseEvent

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**ReleaseId** | Pointer to **string** |  | [optional] 
**StackId** | Pointer to **string** |  | [optional] 
**Sequence** | Pointer to **int32** |  | [optional] 
**OccurredAt** | Pointer to **time.Time** |  | [optional] 
**Source** | Pointer to **string** |  | [optional] 
**Scope** | Pointer to **string** |  | [optional] 
**ResourceName** | Pointer to **string** |  | [optional] 
**Type** | Pointer to **string** |  | [optional] 
**Level** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Links** | Pointer to [**[]ReleaseEventLink**](ReleaseEventLink.md) |  | [optional] 
**Metadata** | Pointer to **map[string]string** |  | [optional] 

## Methods

### NewReleaseEvent

`func NewReleaseEvent() *ReleaseEvent`

NewReleaseEvent instantiates a new ReleaseEvent object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReleaseEventWithDefaults

`func NewReleaseEventWithDefaults() *ReleaseEvent`

NewReleaseEventWithDefaults instantiates a new ReleaseEvent object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ReleaseEvent) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ReleaseEvent) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ReleaseEvent) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ReleaseEvent) HasId() bool`

HasId returns a boolean if a field has been set.

### GetReleaseId

`func (o *ReleaseEvent) GetReleaseId() string`

GetReleaseId returns the ReleaseId field if non-nil, zero value otherwise.

### GetReleaseIdOk

`func (o *ReleaseEvent) GetReleaseIdOk() (*string, bool)`

GetReleaseIdOk returns a tuple with the ReleaseId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReleaseId

`func (o *ReleaseEvent) SetReleaseId(v string)`

SetReleaseId sets ReleaseId field to given value.

### HasReleaseId

`func (o *ReleaseEvent) HasReleaseId() bool`

HasReleaseId returns a boolean if a field has been set.

### GetStackId

`func (o *ReleaseEvent) GetStackId() string`

GetStackId returns the StackId field if non-nil, zero value otherwise.

### GetStackIdOk

`func (o *ReleaseEvent) GetStackIdOk() (*string, bool)`

GetStackIdOk returns a tuple with the StackId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackId

`func (o *ReleaseEvent) SetStackId(v string)`

SetStackId sets StackId field to given value.

### HasStackId

`func (o *ReleaseEvent) HasStackId() bool`

HasStackId returns a boolean if a field has been set.

### GetSequence

`func (o *ReleaseEvent) GetSequence() int32`

GetSequence returns the Sequence field if non-nil, zero value otherwise.

### GetSequenceOk

`func (o *ReleaseEvent) GetSequenceOk() (*int32, bool)`

GetSequenceOk returns a tuple with the Sequence field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSequence

`func (o *ReleaseEvent) SetSequence(v int32)`

SetSequence sets Sequence field to given value.

### HasSequence

`func (o *ReleaseEvent) HasSequence() bool`

HasSequence returns a boolean if a field has been set.

### GetOccurredAt

`func (o *ReleaseEvent) GetOccurredAt() time.Time`

GetOccurredAt returns the OccurredAt field if non-nil, zero value otherwise.

### GetOccurredAtOk

`func (o *ReleaseEvent) GetOccurredAtOk() (*time.Time, bool)`

GetOccurredAtOk returns a tuple with the OccurredAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOccurredAt

`func (o *ReleaseEvent) SetOccurredAt(v time.Time)`

SetOccurredAt sets OccurredAt field to given value.

### HasOccurredAt

`func (o *ReleaseEvent) HasOccurredAt() bool`

HasOccurredAt returns a boolean if a field has been set.

### GetSource

`func (o *ReleaseEvent) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *ReleaseEvent) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *ReleaseEvent) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *ReleaseEvent) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetScope

`func (o *ReleaseEvent) GetScope() string`

GetScope returns the Scope field if non-nil, zero value otherwise.

### GetScopeOk

`func (o *ReleaseEvent) GetScopeOk() (*string, bool)`

GetScopeOk returns a tuple with the Scope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScope

`func (o *ReleaseEvent) SetScope(v string)`

SetScope sets Scope field to given value.

### HasScope

`func (o *ReleaseEvent) HasScope() bool`

HasScope returns a boolean if a field has been set.

### GetResourceName

`func (o *ReleaseEvent) GetResourceName() string`

GetResourceName returns the ResourceName field if non-nil, zero value otherwise.

### GetResourceNameOk

`func (o *ReleaseEvent) GetResourceNameOk() (*string, bool)`

GetResourceNameOk returns a tuple with the ResourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceName

`func (o *ReleaseEvent) SetResourceName(v string)`

SetResourceName sets ResourceName field to given value.

### HasResourceName

`func (o *ReleaseEvent) HasResourceName() bool`

HasResourceName returns a boolean if a field has been set.

### GetType

`func (o *ReleaseEvent) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *ReleaseEvent) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *ReleaseEvent) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *ReleaseEvent) HasType() bool`

HasType returns a boolean if a field has been set.

### GetLevel

`func (o *ReleaseEvent) GetLevel() string`

GetLevel returns the Level field if non-nil, zero value otherwise.

### GetLevelOk

`func (o *ReleaseEvent) GetLevelOk() (*string, bool)`

GetLevelOk returns a tuple with the Level field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLevel

`func (o *ReleaseEvent) SetLevel(v string)`

SetLevel sets Level field to given value.

### HasLevel

`func (o *ReleaseEvent) HasLevel() bool`

HasLevel returns a boolean if a field has been set.

### GetMessage

`func (o *ReleaseEvent) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ReleaseEvent) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ReleaseEvent) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ReleaseEvent) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetLinks

`func (o *ReleaseEvent) GetLinks() []ReleaseEventLink`

GetLinks returns the Links field if non-nil, zero value otherwise.

### GetLinksOk

`func (o *ReleaseEvent) GetLinksOk() (*[]ReleaseEventLink, bool)`

GetLinksOk returns a tuple with the Links field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLinks

`func (o *ReleaseEvent) SetLinks(v []ReleaseEventLink)`

SetLinks sets Links field to given value.

### HasLinks

`func (o *ReleaseEvent) HasLinks() bool`

HasLinks returns a boolean if a field has been set.

### GetMetadata

`func (o *ReleaseEvent) GetMetadata() map[string]string`

GetMetadata returns the Metadata field if non-nil, zero value otherwise.

### GetMetadataOk

`func (o *ReleaseEvent) GetMetadataOk() (*map[string]string, bool)`

GetMetadataOk returns a tuple with the Metadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetadata

`func (o *ReleaseEvent) SetMetadata(v map[string]string)`

SetMetadata sets Metadata field to given value.

### HasMetadata

`func (o *ReleaseEvent) HasMetadata() bool`

HasMetadata returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


