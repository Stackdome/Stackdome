# WorkspaceResource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**WorkspaceId** | Pointer to **string** |  | [optional] [readonly] 
**Name** | **string** |  | 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 
**Version** | Pointer to **int32** |  | [optional] [readonly] 
**ImageRegistry** | Pointer to **string** |  | [optional] 
**Build** | Pointer to [**BuildConfig**](BuildConfig.md) |  | [optional] 
**Prebuilt** | Pointer to [**PrebuiltConfig**](PrebuiltConfig.md) |  | [optional] 
**Init** | Pointer to [**InitConfig**](InitConfig.md) |  | [optional] 
**ExecutionConfig** | Pointer to [**ExecutionConfig**](ExecutionConfig.md) |  | [optional] 
**VolumeMounts** | Pointer to [**[]VolumeMount**](VolumeMount.md) |  | [optional] 
**DependsOn** | Pointer to **[]string** |  | [optional] 
**LifecycleConfig** | Pointer to [**LifecycleConfig**](LifecycleConfig.md) |  | [optional] 
**Ports** | Pointer to [**[]Port**](Port.md) |  | [optional] 
**Stateful** | Pointer to **bool** |  | [optional] 
**Status** | Pointer to [**ResourceStatus**](ResourceStatus.md) |  | [optional] 

## Methods

### NewWorkspaceResource

`func NewWorkspaceResource(name string, ) *WorkspaceResource`

NewWorkspaceResource instantiates a new WorkspaceResource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceResourceWithDefaults

`func NewWorkspaceResourceWithDefaults() *WorkspaceResource`

NewWorkspaceResourceWithDefaults instantiates a new WorkspaceResource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkspaceResource) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkspaceResource) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkspaceResource) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *WorkspaceResource) HasId() bool`

HasId returns a boolean if a field has been set.

### GetWorkspaceId

`func (o *WorkspaceResource) GetWorkspaceId() string`

GetWorkspaceId returns the WorkspaceId field if non-nil, zero value otherwise.

### GetWorkspaceIdOk

`func (o *WorkspaceResource) GetWorkspaceIdOk() (*string, bool)`

GetWorkspaceIdOk returns a tuple with the WorkspaceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceId

`func (o *WorkspaceResource) SetWorkspaceId(v string)`

SetWorkspaceId sets WorkspaceId field to given value.

### HasWorkspaceId

`func (o *WorkspaceResource) HasWorkspaceId() bool`

HasWorkspaceId returns a boolean if a field has been set.

### GetName

`func (o *WorkspaceResource) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkspaceResource) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkspaceResource) SetName(v string)`

SetName sets Name field to given value.


### GetLabels

`func (o *WorkspaceResource) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *WorkspaceResource) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *WorkspaceResource) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *WorkspaceResource) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *WorkspaceResource) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *WorkspaceResource) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *WorkspaceResource) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *WorkspaceResource) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetVersion

`func (o *WorkspaceResource) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkspaceResource) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkspaceResource) SetVersion(v int32)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkspaceResource) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetImageRegistry

`func (o *WorkspaceResource) GetImageRegistry() string`

GetImageRegistry returns the ImageRegistry field if non-nil, zero value otherwise.

### GetImageRegistryOk

`func (o *WorkspaceResource) GetImageRegistryOk() (*string, bool)`

GetImageRegistryOk returns a tuple with the ImageRegistry field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageRegistry

`func (o *WorkspaceResource) SetImageRegistry(v string)`

SetImageRegistry sets ImageRegistry field to given value.

### HasImageRegistry

`func (o *WorkspaceResource) HasImageRegistry() bool`

HasImageRegistry returns a boolean if a field has been set.

### GetBuild

`func (o *WorkspaceResource) GetBuild() BuildConfig`

GetBuild returns the Build field if non-nil, zero value otherwise.

### GetBuildOk

`func (o *WorkspaceResource) GetBuildOk() (*BuildConfig, bool)`

GetBuildOk returns a tuple with the Build field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuild

`func (o *WorkspaceResource) SetBuild(v BuildConfig)`

SetBuild sets Build field to given value.

### HasBuild

`func (o *WorkspaceResource) HasBuild() bool`

HasBuild returns a boolean if a field has been set.

### GetPrebuilt

`func (o *WorkspaceResource) GetPrebuilt() PrebuiltConfig`

GetPrebuilt returns the Prebuilt field if non-nil, zero value otherwise.

### GetPrebuiltOk

`func (o *WorkspaceResource) GetPrebuiltOk() (*PrebuiltConfig, bool)`

GetPrebuiltOk returns a tuple with the Prebuilt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrebuilt

`func (o *WorkspaceResource) SetPrebuilt(v PrebuiltConfig)`

SetPrebuilt sets Prebuilt field to given value.

### HasPrebuilt

`func (o *WorkspaceResource) HasPrebuilt() bool`

HasPrebuilt returns a boolean if a field has been set.

### GetInit

`func (o *WorkspaceResource) GetInit() InitConfig`

GetInit returns the Init field if non-nil, zero value otherwise.

### GetInitOk

`func (o *WorkspaceResource) GetInitOk() (*InitConfig, bool)`

GetInitOk returns a tuple with the Init field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInit

`func (o *WorkspaceResource) SetInit(v InitConfig)`

SetInit sets Init field to given value.

### HasInit

`func (o *WorkspaceResource) HasInit() bool`

HasInit returns a boolean if a field has been set.

### GetExecutionConfig

`func (o *WorkspaceResource) GetExecutionConfig() ExecutionConfig`

GetExecutionConfig returns the ExecutionConfig field if non-nil, zero value otherwise.

### GetExecutionConfigOk

`func (o *WorkspaceResource) GetExecutionConfigOk() (*ExecutionConfig, bool)`

GetExecutionConfigOk returns a tuple with the ExecutionConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExecutionConfig

`func (o *WorkspaceResource) SetExecutionConfig(v ExecutionConfig)`

SetExecutionConfig sets ExecutionConfig field to given value.

### HasExecutionConfig

`func (o *WorkspaceResource) HasExecutionConfig() bool`

HasExecutionConfig returns a boolean if a field has been set.

### GetVolumeMounts

`func (o *WorkspaceResource) GetVolumeMounts() []VolumeMount`

GetVolumeMounts returns the VolumeMounts field if non-nil, zero value otherwise.

### GetVolumeMountsOk

`func (o *WorkspaceResource) GetVolumeMountsOk() (*[]VolumeMount, bool)`

GetVolumeMountsOk returns a tuple with the VolumeMounts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeMounts

`func (o *WorkspaceResource) SetVolumeMounts(v []VolumeMount)`

SetVolumeMounts sets VolumeMounts field to given value.

### HasVolumeMounts

`func (o *WorkspaceResource) HasVolumeMounts() bool`

HasVolumeMounts returns a boolean if a field has been set.

### GetDependsOn

`func (o *WorkspaceResource) GetDependsOn() []string`

GetDependsOn returns the DependsOn field if non-nil, zero value otherwise.

### GetDependsOnOk

`func (o *WorkspaceResource) GetDependsOnOk() (*[]string, bool)`

GetDependsOnOk returns a tuple with the DependsOn field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependsOn

`func (o *WorkspaceResource) SetDependsOn(v []string)`

SetDependsOn sets DependsOn field to given value.

### HasDependsOn

`func (o *WorkspaceResource) HasDependsOn() bool`

HasDependsOn returns a boolean if a field has been set.

### GetLifecycleConfig

`func (o *WorkspaceResource) GetLifecycleConfig() LifecycleConfig`

GetLifecycleConfig returns the LifecycleConfig field if non-nil, zero value otherwise.

### GetLifecycleConfigOk

`func (o *WorkspaceResource) GetLifecycleConfigOk() (*LifecycleConfig, bool)`

GetLifecycleConfigOk returns a tuple with the LifecycleConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLifecycleConfig

`func (o *WorkspaceResource) SetLifecycleConfig(v LifecycleConfig)`

SetLifecycleConfig sets LifecycleConfig field to given value.

### HasLifecycleConfig

`func (o *WorkspaceResource) HasLifecycleConfig() bool`

HasLifecycleConfig returns a boolean if a field has been set.

### GetPorts

`func (o *WorkspaceResource) GetPorts() []Port`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *WorkspaceResource) GetPortsOk() (*[]Port, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *WorkspaceResource) SetPorts(v []Port)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *WorkspaceResource) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### GetStateful

`func (o *WorkspaceResource) GetStateful() bool`

GetStateful returns the Stateful field if non-nil, zero value otherwise.

### GetStatefulOk

`func (o *WorkspaceResource) GetStatefulOk() (*bool, bool)`

GetStatefulOk returns a tuple with the Stateful field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateful

`func (o *WorkspaceResource) SetStateful(v bool)`

SetStateful sets Stateful field to given value.

### HasStateful

`func (o *WorkspaceResource) HasStateful() bool`

HasStateful returns a boolean if a field has been set.

### GetStatus

`func (o *WorkspaceResource) GetStatus() ResourceStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *WorkspaceResource) GetStatusOk() (*ResourceStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *WorkspaceResource) SetStatus(v ResourceStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *WorkspaceResource) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


