# ReleaseOutcome

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | Pointer to [**map[string]ResourceOutcome**](ResourceOutcome.md) |  | [optional] 
**Duration** | Pointer to **string** |  | [optional] 

## Methods

### NewReleaseOutcome

`func NewReleaseOutcome() *ReleaseOutcome`

NewReleaseOutcome instantiates a new ReleaseOutcome object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReleaseOutcomeWithDefaults

`func NewReleaseOutcomeWithDefaults() *ReleaseOutcome`

NewReleaseOutcomeWithDefaults instantiates a new ReleaseOutcome object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *ReleaseOutcome) GetResources() map[string]ResourceOutcome`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *ReleaseOutcome) GetResourcesOk() (*map[string]ResourceOutcome, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *ReleaseOutcome) SetResources(v map[string]ResourceOutcome)`

SetResources sets Resources field to given value.

### HasResources

`func (o *ReleaseOutcome) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetDuration

`func (o *ReleaseOutcome) GetDuration() string`

GetDuration returns the Duration field if non-nil, zero value otherwise.

### GetDurationOk

`func (o *ReleaseOutcome) GetDurationOk() (*string, bool)`

GetDurationOk returns a tuple with the Duration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDuration

`func (o *ReleaseOutcome) SetDuration(v string)`

SetDuration sets Duration field to given value.

### HasDuration

`func (o *ReleaseOutcome) HasDuration() bool`

HasDuration returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


