# RollbackReleaseRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**InstalledAppId** | Pointer to **int32** | Installed App Id if the app is installed from chart store | [optional] 
**InstalledAppVersionId** | Pointer to **int32** | Installed App Version Id if the app is installed from chart store | [optional] 
**HAppId** | Pointer to **string** | helm App Id if the application is installed from using helm (for example \&quot;clusterId|namespace|appName\&quot; ) | [optional] 
**Version** | Pointer to **int32** | rollback to this version | [optional] 

## Methods

### NewRollbackReleaseRequest

`func NewRollbackReleaseRequest() *RollbackReleaseRequest`

NewRollbackReleaseRequest instantiates a new RollbackReleaseRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRollbackReleaseRequestWithDefaults

`func NewRollbackReleaseRequestWithDefaults() *RollbackReleaseRequest`

NewRollbackReleaseRequestWithDefaults instantiates a new RollbackReleaseRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInstalledAppId

`func (o *RollbackReleaseRequest) GetInstalledAppId() int32`

GetInstalledAppId returns the InstalledAppId field if non-nil, zero value otherwise.

### GetInstalledAppIdOk

`func (o *RollbackReleaseRequest) GetInstalledAppIdOk() (*int32, bool)`

GetInstalledAppIdOk returns a tuple with the InstalledAppId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstalledAppId

`func (o *RollbackReleaseRequest) SetInstalledAppId(v int32)`

SetInstalledAppId sets InstalledAppId field to given value.

### HasInstalledAppId

`func (o *RollbackReleaseRequest) HasInstalledAppId() bool`

HasInstalledAppId returns a boolean if a field has been set.

### GetInstalledAppVersionId

`func (o *RollbackReleaseRequest) GetInstalledAppVersionId() int32`

GetInstalledAppVersionId returns the InstalledAppVersionId field if non-nil, zero value otherwise.

### GetInstalledAppVersionIdOk

`func (o *RollbackReleaseRequest) GetInstalledAppVersionIdOk() (*int32, bool)`

GetInstalledAppVersionIdOk returns a tuple with the InstalledAppVersionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstalledAppVersionId

`func (o *RollbackReleaseRequest) SetInstalledAppVersionId(v int32)`

SetInstalledAppVersionId sets InstalledAppVersionId field to given value.

### HasInstalledAppVersionId

`func (o *RollbackReleaseRequest) HasInstalledAppVersionId() bool`

HasInstalledAppVersionId returns a boolean if a field has been set.

### GetHAppId

`func (o *RollbackReleaseRequest) GetHAppId() string`

GetHAppId returns the HAppId field if non-nil, zero value otherwise.

### GetHAppIdOk

`func (o *RollbackReleaseRequest) GetHAppIdOk() (*string, bool)`

GetHAppIdOk returns a tuple with the HAppId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHAppId

`func (o *RollbackReleaseRequest) SetHAppId(v string)`

SetHAppId sets HAppId field to given value.

### HasHAppId

`func (o *RollbackReleaseRequest) HasHAppId() bool`

HasHAppId returns a boolean if a field has been set.

### GetVersion

`func (o *RollbackReleaseRequest) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *RollbackReleaseRequest) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *RollbackReleaseRequest) SetVersion(v int32)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *RollbackReleaseRequest) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


