# ApiToken

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **int32** | Id of api-token | [optional] 
**UserId** | Pointer to **int32** | User Id associated with api-token | [optional] 
**UserIdentifier** | Pointer to **string** | EmailId of that api-token user | [optional] 
**Name** | Pointer to **string** | Name of api-token | [optional] 
**Description** | Pointer to **string** | Description of api-token | [optional] 
**ExpireAtInMs** | Pointer to **int64** | Expiration time of api-token in milliseconds | [optional] 
**Token** | Pointer to **string** | Token of that api-token | [optional] 
**LastUsedAt** | Pointer to **string** | Date of Last used of this token | [optional] 
**LastUsedByIp** | Pointer to **string** | token last used by IP | [optional] 
**UpdatedAt** | Pointer to **string** | token last updatedAt | [optional] 

## Methods

### NewApiToken

`func NewApiToken() *ApiToken`

NewApiToken instantiates a new ApiToken object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewApiTokenWithDefaults

`func NewApiTokenWithDefaults() *ApiToken`

NewApiTokenWithDefaults instantiates a new ApiToken object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ApiToken) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ApiToken) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ApiToken) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *ApiToken) HasId() bool`

HasId returns a boolean if a field has been set.

### GetUserId

`func (o *ApiToken) GetUserId() int32`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *ApiToken) GetUserIdOk() (*int32, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *ApiToken) SetUserId(v int32)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *ApiToken) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetUserIdentifier

`func (o *ApiToken) GetUserIdentifier() string`

GetUserIdentifier returns the UserIdentifier field if non-nil, zero value otherwise.

### GetUserIdentifierOk

`func (o *ApiToken) GetUserIdentifierOk() (*string, bool)`

GetUserIdentifierOk returns a tuple with the UserIdentifier field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserIdentifier

`func (o *ApiToken) SetUserIdentifier(v string)`

SetUserIdentifier sets UserIdentifier field to given value.

### HasUserIdentifier

`func (o *ApiToken) HasUserIdentifier() bool`

HasUserIdentifier returns a boolean if a field has been set.

### GetName

`func (o *ApiToken) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ApiToken) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ApiToken) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ApiToken) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDescription

`func (o *ApiToken) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *ApiToken) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *ApiToken) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *ApiToken) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetExpireAtInMs

`func (o *ApiToken) GetExpireAtInMs() int64`

GetExpireAtInMs returns the ExpireAtInMs field if non-nil, zero value otherwise.

### GetExpireAtInMsOk

`func (o *ApiToken) GetExpireAtInMsOk() (*int64, bool)`

GetExpireAtInMsOk returns a tuple with the ExpireAtInMs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpireAtInMs

`func (o *ApiToken) SetExpireAtInMs(v int64)`

SetExpireAtInMs sets ExpireAtInMs field to given value.

### HasExpireAtInMs

`func (o *ApiToken) HasExpireAtInMs() bool`

HasExpireAtInMs returns a boolean if a field has been set.

### GetToken

`func (o *ApiToken) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *ApiToken) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *ApiToken) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *ApiToken) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetLastUsedAt

`func (o *ApiToken) GetLastUsedAt() string`

GetLastUsedAt returns the LastUsedAt field if non-nil, zero value otherwise.

### GetLastUsedAtOk

`func (o *ApiToken) GetLastUsedAtOk() (*string, bool)`

GetLastUsedAtOk returns a tuple with the LastUsedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUsedAt

`func (o *ApiToken) SetLastUsedAt(v string)`

SetLastUsedAt sets LastUsedAt field to given value.

### HasLastUsedAt

`func (o *ApiToken) HasLastUsedAt() bool`

HasLastUsedAt returns a boolean if a field has been set.

### GetLastUsedByIp

`func (o *ApiToken) GetLastUsedByIp() string`

GetLastUsedByIp returns the LastUsedByIp field if non-nil, zero value otherwise.

### GetLastUsedByIpOk

`func (o *ApiToken) GetLastUsedByIpOk() (*string, bool)`

GetLastUsedByIpOk returns a tuple with the LastUsedByIp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastUsedByIp

`func (o *ApiToken) SetLastUsedByIp(v string)`

SetLastUsedByIp sets LastUsedByIp field to given value.

### HasLastUsedByIp

`func (o *ApiToken) HasLastUsedByIp() bool`

HasLastUsedByIp returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *ApiToken) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *ApiToken) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *ApiToken) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *ApiToken) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


