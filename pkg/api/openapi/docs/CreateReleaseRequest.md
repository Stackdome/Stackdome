# CreateReleaseRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FromReleaseId** | Pointer to **string** | If set, creates a rollback release copying this release&#39;s manifest | [optional] 

## Methods

### NewCreateReleaseRequest

`func NewCreateReleaseRequest() *CreateReleaseRequest`

NewCreateReleaseRequest instantiates a new CreateReleaseRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateReleaseRequestWithDefaults

`func NewCreateReleaseRequestWithDefaults() *CreateReleaseRequest`

NewCreateReleaseRequestWithDefaults instantiates a new CreateReleaseRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFromReleaseId

`func (o *CreateReleaseRequest) GetFromReleaseId() string`

GetFromReleaseId returns the FromReleaseId field if non-nil, zero value otherwise.

### GetFromReleaseIdOk

`func (o *CreateReleaseRequest) GetFromReleaseIdOk() (*string, bool)`

GetFromReleaseIdOk returns a tuple with the FromReleaseId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFromReleaseId

`func (o *CreateReleaseRequest) SetFromReleaseId(v string)`

SetFromReleaseId sets FromReleaseId field to given value.

### HasFromReleaseId

`func (o *CreateReleaseRequest) HasFromReleaseId() bool`

HasFromReleaseId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


