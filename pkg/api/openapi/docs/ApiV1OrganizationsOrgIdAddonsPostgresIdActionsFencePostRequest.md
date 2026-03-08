# ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Fence** | **bool** | true to fence the cluster, false to unfence | 
**Reason** | Pointer to **string** | Reason for fencing/unfencing | [optional] 

## Methods

### NewApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest

`func NewApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest(fence bool, ) *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest`

NewApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest instantiates a new ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequestWithDefaults

`func NewApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequestWithDefaults() *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest`

NewApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequestWithDefaults instantiates a new ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFence

`func (o *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest) GetFence() bool`

GetFence returns the Fence field if non-nil, zero value otherwise.

### GetFenceOk

`func (o *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest) GetFenceOk() (*bool, bool)`

GetFenceOk returns a tuple with the Fence field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFence

`func (o *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest) SetFence(v bool)`

SetFence sets Fence field to given value.


### GetReason

`func (o *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest) HasReason() bool`

HasReason returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


