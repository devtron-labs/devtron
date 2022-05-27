# CreateApiTokenRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Name of api-token | [optional] 
**Description** | Pointer to **string** | Description of api-token | [optional] 
**ExpireAtInMs** | Pointer to **int64** | Expiration time of api-token in milliseconds | [optional] 

## Methods

### NewCreateApiTokenRequest

`func NewCreateApiTokenRequest() *CreateApiTokenRequest`

NewCreateApiTokenRequest instantiates a new CreateApiTokenRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateApiTokenRequestWithDefaults

`func NewCreateApiTokenRequestWithDefaults() *CreateApiTokenRequest`

NewCreateApiTokenRequestWithDefaults instantiates a new CreateApiTokenRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *CreateApiTokenRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateApiTokenRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateApiTokenRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *CreateApiTokenRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDescription

`func (o *CreateApiTokenRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *CreateApiTokenRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *CreateApiTokenRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *CreateApiTokenRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetExpireAtInMs

`func (o *CreateApiTokenRequest) GetExpireAtInMs() int64`

GetExpireAtInMs returns the ExpireAtInMs field if non-nil, zero value otherwise.

### GetExpireAtInMsOk

`func (o *CreateApiTokenRequest) GetExpireAtInMsOk() (*int64, bool)`

GetExpireAtInMsOk returns a tuple with the ExpireAtInMs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpireAtInMs

`func (o *CreateApiTokenRequest) SetExpireAtInMs(v int64)`

SetExpireAtInMs sets ExpireAtInMs field to given value.

### HasExpireAtInMs

`func (o *CreateApiTokenRequest) HasExpireAtInMs() bool`

HasExpireAtInMs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


