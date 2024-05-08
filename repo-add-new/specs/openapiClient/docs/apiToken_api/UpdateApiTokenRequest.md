# UpdateApiTokenRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** | Description of api-token | [optional] 
**ExpireAtInMs** | Pointer to **int64** | Expiration time of api-token in milliseconds | [optional] 

## Methods

### NewUpdateApiTokenRequest

`func NewUpdateApiTokenRequest() *UpdateApiTokenRequest`

NewUpdateApiTokenRequest instantiates a new UpdateApiTokenRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateApiTokenRequestWithDefaults

`func NewUpdateApiTokenRequestWithDefaults() *UpdateApiTokenRequest`

NewUpdateApiTokenRequestWithDefaults instantiates a new UpdateApiTokenRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *UpdateApiTokenRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *UpdateApiTokenRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *UpdateApiTokenRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *UpdateApiTokenRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetExpireAtInMs

`func (o *UpdateApiTokenRequest) GetExpireAtInMs() int64`

GetExpireAtInMs returns the ExpireAtInMs field if non-nil, zero value otherwise.

### GetExpireAtInMsOk

`func (o *UpdateApiTokenRequest) GetExpireAtInMsOk() (*int64, bool)`

GetExpireAtInMsOk returns a tuple with the ExpireAtInMs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpireAtInMs

`func (o *UpdateApiTokenRequest) SetExpireAtInMs(v int64)`

SetExpireAtInMs sets ExpireAtInMs field to given value.

### HasExpireAtInMs

`func (o *UpdateApiTokenRequest) HasExpireAtInMs() bool`

HasExpireAtInMs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


