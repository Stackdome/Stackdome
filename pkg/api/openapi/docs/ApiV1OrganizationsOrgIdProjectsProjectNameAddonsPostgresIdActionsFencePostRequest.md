# ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Fence** | **bool** | true to fence the cluster, false to unfence | 
**Reason** | Pointer to **string** | Reason for fencing/unfencing | [optional] 

## Methods

### NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest

`func NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest(fence bool, ) *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest`

NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest instantiates a new ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequestWithDefaults

`func NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequestWithDefaults() *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest`

NewApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequestWithDefaults instantiates a new ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFence

`func (o *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest) GetFence() bool`

GetFence returns the Fence field if non-nil, zero value otherwise.

### GetFenceOk

`func (o *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest) GetFenceOk() (*bool, bool)`

GetFenceOk returns a tuple with the Fence field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFence

`func (o *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest) SetFence(v bool)`

SetFence sets Fence field to given value.


### GetReason

`func (o *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *ApiV1OrganizationsOrgIdProjectsProjectNameAddonsPostgresIdActionsFencePostRequest) HasReason() bool`

HasReason returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


